package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"pitchle/services/ingestion/internal/normalizer"
	"pitchle/services/ingestion/internal/parser"
	"pitchle/services/ingestion/internal/repository"
	"pitchle/shared/models"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pitchle?sslmode=disable"
	}

	csvPath := os.Getenv("SAVANT_CSV_PATH")
	if csvPath == "" {
		csvPath = "savant_data.csv"
	}

	log.Printf("Starting ingestion service...")
	log.Printf("CSV Path: %s", csvPath)
	var db *sql.DB
	var err error
	for i := range 30 {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to database")
				break
			}
		}
		log.Printf("Database connection attempt %d failed: %v. Retrying in 2 seconds...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()
	repo := repository.NewRepository(db)
	var teamMap map[int]int
	for i := range 15 {
		teamMap, err = repo.GetTeamMap()
		if err == nil && len(teamMap) > 0 {
			break
		}
		log.Printf("Waiting for teams seed data to be available... attempt %d: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil || len(teamMap) == 0 {
		log.Fatalf("Failed to fetch teams: %v (team map size: %d)", err, len(teamMap))
	}

	log.Printf("Loaded team map with %d teams", len(teamMap))

	// Find the absolute CSV path
	absCSVPath, err := filepath.Abs(csvPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute CSV path: %v", err)
	}

	records, err := parser.ParseCSV(absCSVPath)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	log.Printf("Parsed %d records from CSV", len(records))

	metadataPath := os.Getenv("PLAYER_METADATA_CSV_PATH")
	if metadataPath == "" {
		metadataPath = "player_metadata.csv"
	}

	absMetadataPath, err := filepath.Abs(metadataPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute metadata CSV path: %v", err)
	}

	metadataMap, err := parser.ParseMetadataCSV(absMetadataPath)
	if err != nil {
		log.Fatalf("Failed to parse metadata CSV: %v", err)
	}
	log.Printf("Successfully parsed %d metadata records from %s", len(metadataMap), absMetadataPath)

	playerMap := make(map[int]models.Player)
	for pid, meta := range metadataMap {
		var birthDate *time.Time
		if meta.BirthDate != "" {
			t, err := time.Parse("2006-01-02", meta.BirthDate)
			if err == nil {
				birthDate = &t
			}
		}

		var dbTeamID int
		if id, ok := teamMap[meta.MLBTeamID]; ok {
			dbTeamID = id
		}
		if dbTeamID == 0 {
			// Fallback team mapping using deterministic fallback
			mlbTeamID := normalizer.MLBTeamIDs[pid%len(normalizer.MLBTeamIDs)]
			dbTeamID = teamMap[mlbTeamID]
		}

		playerMap[pid] = models.Player{
			MLBID:        pid,
			Name:         meta.PlayerName,
			BirthDate:    birthDate,
			BirthYear:    meta.BirthYear,
			BirthCity:    meta.BirthCity,
			BirthCountry: meta.BirthCountry,
			Position:     meta.Position,
			Height:       meta.Height,
			Weight:       meta.Weight,
			MLBDebutYear: meta.MLBDebutYear,
			MLBLastYear:  meta.MLBLastYear,
			TeamID:       dbTeamID,
		}
	}

	type jobResult struct {
		success  bool
		playerID int
		name     string
		err      error
	}

	jobs := make(chan parser.PitchRecord, len(records))
	results := make(chan jobResult, len(records))

	numWorkers := 20
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rec := range jobs {
				normalizedName := normalizer.NormalizeName(rec.PlayerName)
				player, ok := playerMap[rec.PlayerID]
				if !ok {
					// Fallback to generating deterministic metadata
					var err error
					player, err = normalizer.GenerateDeterministicMetadata(rec.PlayerID, teamMap)
					if err != nil {
						results <- jobResult{success: false, playerID: rec.PlayerID, name: normalizedName, err: fmt.Errorf("fallback metadata: %w", err)}
						continue
					}
					player.Name = normalizedName
				}

				profile := models.PitchProfile{
					PitchType:        "Four-Seam Fastball",
					Velocity:         rec.Velocity,
					SpinRate:         rec.SpinRate,
					ReleasePosX:      rec.ReleasePosX,
					ReleasePosZ:      rec.ReleasePosZ,
					ReleaseExtension: rec.ReleaseExtension,
					BreakX:           rec.BreakX,
					BreakZ:           rec.BreakZ,
					ArmAngle:         rec.ArmAngle,
					PlateX:           rec.PlateX,
					PlateZ:           rec.PlateZ,
				}

				err = repo.UpsertPlayerAndProfile(player, player.Name, profile)
				if err != nil {
					results <- jobResult{success: false, playerID: rec.PlayerID, name: player.Name, err: fmt.Errorf("upsert: %w", err)}
					continue
				}
				results <- jobResult{success: true, playerID: rec.PlayerID, name: player.Name}
			}
		}()
	}

	// Feed jobs
	for _, rec := range records {
		jobs <- rec
	}
	close(jobs)

	// Wait for workers and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	for res := range results {
		if res.success {
			successCount++
		} else {
			log.Printf("Warning: failed to ingest player %d (%s): %v", res.playerID, res.name, res.err)
		}
	}
	log.Printf("Ingested %d of %d players and pitch profiles successfully", successCount, len(records))
}
