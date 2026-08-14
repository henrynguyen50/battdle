package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"pitchle/services/pitch/internal/physics"
	"pitchle/services/pitch/internal/repository"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pitchle?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Starting pitch service on port %s...", port)

	// Connect to database with retry
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

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /api/v1/puzzle/today/animation", func(w http.ResponseWriter, r *http.Request) {
		// 1. Get today's pitch profile ID
		profileID, err := repo.GetTodayPitchProfileID(time.Now().UTC())
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "daily puzzle not initialized yet. please access today's puzzle first.")
				return
			}
			log.Printf("Error getting today's pitch profile ID: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to get today's pitch profile")
			return
		}

		// 2. Check if animation is already generated and cached
		animData, err := repo.GetAnimationByProfileID(profileID)
		if err == nil {
			// Cache Hit: return directly
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(animData))
			return
		}

		if err != sql.ErrNoRows {
			log.Printf("Error checking animation cache: %v", err)
			// don't fail, proceed to generate
		}

		// Cache Miss: Generate trajectory
		profile, err := repo.GetPitchProfileByID(profileID)
		if err != nil {
			log.Printf("Error getting pitch profile details: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load pitch profile")
			return
		}

		points := physics.CalculateTrajectory(profile)

		// Serialize to JSON
		jsonBytes, err := json.Marshal(points)
		if err != nil {
			log.Printf("Error marshalling trajectory points: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to generate animation data")
			return
		}

		jsonStr := string(jsonBytes)

		// Cache in database
		err = repo.SaveAnimation(profileID, jsonStr)
		if err != nil {
			log.Printf("Warning: failed to cache animation: %v", err)
			// don't fail, return result anyway
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonBytes)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving HTTP on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
