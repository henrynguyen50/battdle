package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"pitchle/services/game/internal/repository"
	"pitchle/services/game/internal/service"
	"pitchle/shared/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pitchle?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting game service on port %s...", port)

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
	gameSvc := service.NewGameService(repo)
	puzzleSvc := service.NewPuzzleService(repo)

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /api/v1/players/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		players, err := repo.SearchPlayers(q)
		if err != nil {
			log.Printf("Error searching players: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to search players")
			return
		}
		if players == nil {
			players = []models.Player{} // ensure non-null array response
		}
		// Map to [{id, name}] format
		type searchResult struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		results := make([]searchResult, len(players))
		for i, p := range players {
			results[i] = searchResult{ID: p.ID, Name: p.Name}
		}
		writeJSON(w, http.StatusOK, results)
	})

	mux.HandleFunc("GET /api/v1/puzzle/today", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = "anonymous"
		}

		puzzle, err := puzzleSvc.GetTodayPuzzle()
		if err != nil {
			log.Printf("Error getting today's puzzle: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily puzzle")
			return
		}

		state, err := gameSvc.GetGameState(sessionID, puzzle)
		if err != nil {
			log.Printf("Error getting game state: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load game state")
			return
		}

		writeJSON(w, http.StatusOK, state)
	})

	mux.HandleFunc("POST /api/v1/puzzle/today/guess", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "missing X-Session-ID header")
			return
		}

		var req struct {
			PlayerID int `json:"player_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.PlayerID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid player_id")
			return
		}

		puzzle, err := puzzleSvc.GetTodayPuzzle()
		if err != nil {
			log.Printf("Error getting today's puzzle: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily puzzle")
			return
		}

		state, err := gameSvc.SubmitGuess(sessionID, puzzle, req.PlayerID)
		if err != nil {
			log.Printf("Error submitting guess: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, state)
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
