package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gameURLStr := os.Getenv("GAME_SERVICE_URL")
	if gameURLStr == "" {
		gameURLStr = "http://localhost:8081"
	}

	pitchURLStr := os.Getenv("PITCH_SERVICE_URL")
	if pitchURLStr == "" {
		pitchURLStr = "http://localhost:8082"
	}

	log.Printf("Starting API Gateway on port %s...", port)
	log.Printf("Routing game service to %s", gameURLStr)
	log.Printf("Routing pitch service to %s", pitchURLStr)

	gameTarget, err := url.Parse(gameURLStr)
	if err != nil {
		log.Fatalf("Invalid game service URL: %v", err)
	}

	pitchTarget, err := url.Parse(pitchURLStr)
	if err != nil {
		log.Fatalf("Invalid pitch service URL: %v", err)
	}

	gameProxy := httputil.NewSingleHostReverseProxy(gameTarget)
	pitchProxy := httputil.NewSingleHostReverseProxy(pitchTarget)

	// Custom Director for Game Proxy to pass headers
	originalGameDirector := gameProxy.Director
	gameProxy.Director = func(req *http.Request) {
		originalGameDirector(req)
		req.Host = gameTarget.Host
	}

	// Custom Director for Pitch Proxy to pass headers
	originalPitchDirector := pitchProxy.Director
	pitchProxy.Director = func(req *http.Request) {
		originalPitchDirector(req)
		req.Host = pitchTarget.Host
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// CORS Middleware
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Handle Session ID
		sessionID := r.Header.Get("X-Session-ID")
		var sessionCookie *http.Cookie
		if sessionID == "" {
			cookie, err := r.Cookie("session_id")
			if err == nil {
				sessionID = cookie.Value
			} else {
				// Generate new session ID (32-character hex)
				bytes := make([]byte, 16)
				_, _ = rand.Read(bytes)
				sessionID = hex.EncodeToString(bytes)
				sessionCookie = &http.Cookie{
					Name:     "session_id",
					Value:    sessionID,
					Path:     "/",
					HttpOnly: false, // Allow reading by frontend JS if needed, but safe
					SameSite: http.SameSiteLaxMode,
					MaxAge:   86400 * 365, // 1 year
				}
			}
		}

		// Add session ID to headers for downstream microservices
		r.Header.Set("X-Session-ID", sessionID)

		// Set cookie in response if generated
		if sessionCookie != nil {
			http.SetCookie(w, sessionCookie)
		}

		// Route Requests
		path := r.URL.Path
		log.Printf("Route: %s %s [Session: %s]", r.Method, path, sessionID)

		if path == "/api/v1/puzzle/today/animation" {
			pitchProxy.ServeHTTP(w, r)
		} else if path == "/api/v1/puzzle/today" || path == "/api/v1/puzzle/today/guess" || path == "/api/v1/puzzle/today/guess-pitch" || path == "/api/v1/players/search" || path == "/api/v1/puzzle/test/reset" || path == "/api/v1/puzzle/test/answer" || path == "/api/v1/puzzle/test/set-pitcher" {
			gameProxy.ServeHTTP(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	server := &http.Server{
		Addr: ":" + port,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Gateway server failed: %v", err)
	}
}
