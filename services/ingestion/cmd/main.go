package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
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

	updatedCSVPath := os.Getenv("UPDATED_CSV_PATH")
	if updatedCSVPath == "" {
		if _, err := os.Stat("updated1.csv"); err == nil {
			updatedCSVPath = "updated1.csv"
		} else if _, err := os.Stat("updated.csv"); err == nil {
			updatedCSVPath = "updated.csv"
		}
	}
	legacyCSVPath := os.Getenv("SAVANT_CSV_PATH")
	if legacyCSVPath == "" {
		legacyCSVPath = "savant_data.csv"
	}

	log.Printf("Starting ingestion service...")
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

	// Check if updated.csv is available
	if updatedCSVPath != "" {
		absUpdatedPath, err := filepath.Abs(updatedCSVPath)
		if err != nil {
			log.Fatalf("Failed to resolve absolute updated CSV path: %v", err)
		}

		log.Printf("Parsing Statcast updated CSV: %s", absUpdatedPath)
		statcastRecords, err := parser.ParseUpdatedCSV(absUpdatedPath)
		if err != nil {
			log.Fatalf("Failed to parse updated CSV: %v", err)
		}
		log.Printf("Parsed %d player Statcast records from %s", len(statcastRecords), absUpdatedPath)

		// Collect player IDs for batch MLB Stats API fetch
		var playerIDs []int
		for _, rec := range statcastRecords {
			playerIDs = append(playerIDs, rec.PlayerID)
		}

		log.Printf("Fetching 2026 MLB team metadata for %d players from MLB Stats API...", len(playerIDs))
		metadataCache, err := normalizer.FetchAllPlayerMetadata(playerIDs, teamMap)
		if err != nil {
			log.Printf("Warning: MLB Stats API batch fetch failed (%v), using fallback metadata", err)
			metadataCache = make(map[int]models.Player)
		} else {
			log.Printf("Successfully fetched metadata for %d players from MLB Stats API", len(metadataCache))
		}

		successCount := 0
		totalPitches := 0
		for _, rec := range statcastRecords {
			meta, ok := metadataCache[rec.PlayerID]
			if !ok {
				meta, _ = normalizer.GenerateDeterministicMetadata(rec.PlayerID, teamMap)
			}

			birthYear := meta.BirthYear
			if birthYear == 0 && rec.PlayerAge > 0 {
				birthYear = 2026 - rec.PlayerAge
			}
			debutYear := meta.MLBDebutYear
			if debutYear == 0 {
				debutYear = birthYear + 22
			}
			lastYear := meta.MLBLastYear
			if lastYear == 0 {
				lastYear = 2026
			}
			pos := meta.Position
			if pos == "" {
				pos = "P"
			}
			teamID := meta.TeamID
			if teamID == 0 {
				mlbTeamID := normalizer.MLBTeamIDs[rec.PlayerID%len(normalizer.MLBTeamIDs)]
				teamID = teamMap[mlbTeamID]
			}

			player := models.Player{
				MLBID:              rec.PlayerID,
				Name:               rec.NormalizedName,
				BirthDate:          meta.BirthDate,
				BirthYear:          birthYear,
				BirthCity:          meta.BirthCity,
				BirthCountry:       meta.BirthCountry,
				Position:           pos,
				Height:             meta.Height,
				Weight:             meta.Weight,
				MLBDebutYear:       debutYear,
				MLBLastYear:        lastYear,
				TeamID:             teamID,
				KPercent:           rec.KPercent,
				BBPercent:          rec.BBPercent,
				InZonePercent:      rec.InZonePercent,
				WhiffPercent:       rec.WhiffPercent,
				GroundballsPercent: rec.GroundballsPercent,
				FlyballsPercent:    rec.FlyballsPercent,
				PopupsPercent:      rec.PopupsPercent,
				PitchHand:          rec.PitchHand,
				ArmAngle:           rec.ArmAngle,
			}

			var profiles []models.PitchProfile
			for _, p := range rec.Pitches {
				profiles = append(profiles, models.PitchProfile{
					PitchType:        p.PitchType,
					Velocity:         p.Velocity,
					SpinRate:         p.SpinRate,
					ReleasePosX:      p.ReleasePosX,
					ReleasePosZ:      p.ReleasePosZ,
					ReleaseExtension: p.ReleaseExtension,
					BreakX:           p.BreakX,
					BreakZ:           p.BreakZ,
					ArmAngle:         p.ArmAngle,
					PlateX:           p.PlateX,
					PlateZ:           p.PlateZ,
					UsagePercent:     p.UsagePercent,
					BreakZInduced:    p.BreakZInduced,
					RangeSpeed:       p.RangeSpeed,
				})
			}

			err := repo.UpsertPlayerWithProfiles(player, rec.NormalizedName, profiles)
			if err != nil {
				log.Printf("Failed to upsert player %s (ID %d): %v", rec.NormalizedName, rec.PlayerID, err)
			} else {
				successCount++
				totalPitches += len(profiles)
			}
		}

		log.Printf("Successfully ingested %d of %d Statcast players (%d pitch profiles)", successCount, len(statcastRecords), totalPitches)
		return
	}

	// Fallback to legacy CSV if updated.csv is not present
	absCSVPath, err := filepath.Abs(legacyCSVPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute CSV path: %v", err)
	}

	records, err := parser.ParseCSV(absCSVPath)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}
	log.Printf("Parsed %d records from legacy CSV", len(records))

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

	for _, rec := range records {
		normalizedName := normalizer.NormalizeName(rec.PlayerName)
		player, ok := playerMap[rec.PlayerID]
		if !ok {
			player, _ = normalizer.GenerateDeterministicMetadata(rec.PlayerID, teamMap)
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
		_ = repo.UpsertPlayerAndProfile(player, player.Name, profile)
	}
}
