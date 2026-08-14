package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	mux.HandleFunc("GET /api/v1/puzzle/today/stats", func(w http.ResponseWriter, r *http.Request) {
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

		stats, err := repo.GetTodayPuzzleStats(puzzle.ID, sessionID)
		if err != nil {
			log.Printf("Error getting today puzzle stats: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load puzzle stats")
			return
		}

		writeJSON(w, http.StatusOK, stats)
	})

	mux.HandleFunc("GET /api/v1/leaderboard/daily", func(w http.ResponseWriter, r *http.Request) {
		puzzle, err := puzzleSvc.GetTodayPuzzle()
		if err != nil {
			log.Printf("Error getting today's puzzle for leaderboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily puzzle")
			return
		}

		limit := 10
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		leaderboard, err := repo.GetDailyLeaderboard(puzzle.ID, limit)
		if err != nil {
			log.Printf("Error getting daily leaderboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily leaderboard")
			return
		}
		if leaderboard == nil {
			leaderboard = []models.LeaderboardEntry{}
		}

		writeJSON(w, http.StatusOK, leaderboard)
	})

	mux.HandleFunc("GET /api/v1/leaderboard/streaks", func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		leaderboard, err := repo.GetStreakLeaderboard(limit)
		if err != nil {
			log.Printf("Error getting streak leaderboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load streak leaderboard")
			return
		}
		if leaderboard == nil {
			leaderboard = []models.StreakLeaderboardEntry{}
		}

		writeJSON(w, http.StatusOK, leaderboard)
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

	mux.HandleFunc("POST /api/v1/puzzle/today/guess-pitch", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "missing X-Session-ID header")
			return
		}

		var req struct {
			PitchType string `json:"pitch_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.PitchType) == "" {
			writeError(w, http.StatusBadRequest, "pitch_type cannot be empty")
			return
		}

		puzzle, err := puzzleSvc.GetTodayPuzzle()
		if err != nil {
			log.Printf("Error getting today's puzzle: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily puzzle")
			return
		}

		state, err := gameSvc.SubmitPitchGuess(sessionID, puzzle, req.PitchType)
		if err != nil {
			log.Printf("Error submitting pitch guess: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, state)
	})

	mux.HandleFunc("POST /api/v1/puzzle/test/reset", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = "anonymous"
		}

		puzzle, err := puzzleSvc.ResetTodayPuzzleForTest(sessionID)
		if err != nil {
			log.Printf("Error resetting test puzzle: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to reset puzzle")
			return
		}

		state, err := gameSvc.GetGameState(sessionID, puzzle)
		if err != nil {
			log.Printf("Error getting game state after reset: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load game state")
			return
		}

		writeJSON(w, http.StatusOK, state)
	})
	mux.HandleFunc("POST /api/v1/puzzle/test/set-pitcher", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = "anonymous"
		}

		var req struct {
			PlayerID int `json:"player_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid player_id")
			return
		}

		puzzle, err := puzzleSvc.SetTargetPitcherForTest(req.PlayerID, sessionID)
		if err != nil {
			log.Printf("Error setting test pitcher: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to set target pitcher")
			return
		}

		state, err := gameSvc.GetGameState(sessionID, puzzle)
		if err != nil {
			log.Printf("Error getting game state after set pitcher: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load game state")
			return
		}

		writeJSON(w, http.StatusOK, state)
	})

	mux.HandleFunc("GET /api/v1/puzzle/test/answer", func(w http.ResponseWriter, r *http.Request) {
		puzzle, err := puzzleSvc.GetTodayPuzzle()
		if err != nil {
			log.Printf("Error getting today's puzzle: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load daily puzzle")
			return
		}

		answer, err := gameSvc.GetPuzzleAnswer(puzzle)
		if err != nil {
			log.Printf("Error getting puzzle answer: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to load puzzle answer")
			return
		}

		writeJSON(w, http.StatusOK, answer)
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
