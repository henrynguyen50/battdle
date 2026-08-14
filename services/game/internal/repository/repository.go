package repository

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"pitchle/shared/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetPlayerByID fetches a player by their database ID, joining team and division info.
func (r *Repository) GetPlayerByID(id int) (*models.Player, error) {
	query := `
		SELECT p.id, p.mlb_id, p.name, p.birth_date, p.birth_year, p.position, p.mlb_debut_year, p.mlb_last_year, p.team_id,
		       COALESCE(t.name, '') AS team_name, COALESCE(t.division_id, 0) AS division_id, COALESCE(d.name, '') AS division_name,
		       COALESCE(d.league, '') AS league,
		       COALESCE(p.k_percent, 0.0), COALESCE(p.bb_percent, 0.0), COALESCE(p.in_zone_percent, 0.0),
		       COALESCE(p.whiff_percent, 0.0), COALESCE(p.groundballs_percent, 0.0), COALESCE(p.flyballs_percent, 0.0),
		       COALESCE(p.popups_percent, 0.0), COALESCE(p.pitch_hand, 'R'), COALESCE(p.arm_angle, 0.0),
		       COALESCE(p.height, ''), COALESCE(p.birth_country, ''), COALESCE(p.birth_city, '')
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN divisions d ON t.division_id = d.id
		WHERE p.id = $1
	`
	var p models.Player
	var birthDate sql.NullTime
	var teamID sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.MLBID, &p.Name, &birthDate, &p.BirthYear, &p.Position, &p.MLBDebutYear, &p.MLBLastYear, &teamID,
		&p.TeamName, &p.DivisionID, &p.DivisionName, &p.League,
		&p.KPercent, &p.BBPercent, &p.InZonePercent,
		&p.WhiffPercent, &p.GroundballsPercent, &p.FlyballsPercent,
		&p.PopupsPercent, &p.PitchHand, &p.ArmAngle,
		&p.Height, &p.BirthCountry, &p.BirthCity,
	)
	if err != nil {
		return nil, err
	}
	if birthDate.Valid {
		p.BirthDate = &birthDate.Time
	}
	if teamID.Valid {
		p.TeamID = int(teamID.Int64)
	}
	return &p, nil
}

// GetPitchProfileByID fetches a pitch profile by its ID.
func (r *Repository) GetPitchProfileByID(id int) (*models.PitchProfile, error) {
	query := `
		SELECT id, player_id, pitch_type, velocity, spin_rate, release_pos_x, release_pos_z,
		       release_extension, break_x, break_z, arm_angle, plate_x, plate_z,
		       COALESCE(usage_percent, 0.0), COALESCE(break_z_induced, 0.0), COALESCE(range_speed, 0.0),
		       created_at
		FROM pitch_profiles
		WHERE id = $1
	`
	var p models.PitchProfile
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.PlayerID, &p.PitchType, &p.Velocity, &p.SpinRate, &p.ReleasePosX, &p.ReleasePosZ,
		&p.ReleaseExtension, &p.BreakX, &p.BreakZ, &p.ArmAngle, &p.PlateX, &p.PlateZ,
		&p.UsagePercent, &p.BreakZInduced, &p.RangeSpeed,
		&p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPitchProfilesByPlayerID fetches all pitch profiles for a player ordered by usage percentage descending.
func (r *Repository) GetPitchProfilesByPlayerID(playerID int) ([]models.PitchProfile, error) {
	query := `
		SELECT id, player_id, pitch_type, velocity, spin_rate, release_pos_x, release_pos_z,
		       release_extension, break_x, break_z, arm_angle, plate_x, plate_z,
		       COALESCE(usage_percent, 0.0), COALESCE(break_z_induced, 0.0), COALESCE(range_speed, 0.0),
		       created_at
		FROM pitch_profiles
		WHERE player_id = $1
		ORDER BY usage_percent DESC, velocity DESC
	`
	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []models.PitchProfile
	for rows.Next() {
		var p models.PitchProfile
		err := rows.Scan(
			&p.ID, &p.PlayerID, &p.PitchType, &p.Velocity, &p.SpinRate, &p.ReleasePosX, &p.ReleasePosZ,
			&p.ReleaseExtension, &p.BreakX, &p.BreakZ, &p.ArmAngle, &p.PlateX, &p.PlateZ,
			&p.UsagePercent, &p.BreakZInduced, &p.RangeSpeed,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// SearchPlayers searches for players by name (case-insensitive substring) returning up to 10 results.
func (r *Repository) SearchPlayers(query string) ([]models.Player, error) {
	sqlQuery := `
		SELECT p.id, p.mlb_id, p.name, p.birth_date, p.birth_year, p.position, p.mlb_debut_year, p.mlb_last_year, p.team_id,
		       COALESCE(t.name, '') AS team_name, COALESCE(t.division_id, 0) AS division_id, COALESCE(d.name, '') AS division_name,
		       COALESCE(d.league, '') AS league,
		       COALESCE(p.k_percent, 0.0), COALESCE(p.bb_percent, 0.0), COALESCE(p.in_zone_percent, 0.0),
		       COALESCE(p.whiff_percent, 0.0), COALESCE(p.groundballs_percent, 0.0), COALESCE(p.flyballs_percent, 0.0),
		       COALESCE(p.popups_percent, 0.0), COALESCE(p.pitch_hand, 'R'), COALESCE(p.arm_angle, 0.0)
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN divisions d ON t.division_id = d.id
		WHERE p.name ILIKE $1
		ORDER BY p.name ASC
		LIMIT 10
	`
	rows, err := r.db.Query(sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var p models.Player
		var birthDate sql.NullTime
		var teamID sql.NullInt64
		err := rows.Scan(
			&p.ID, &p.MLBID, &p.Name, &birthDate, &p.BirthYear, &p.Position, &p.MLBDebutYear, &p.MLBLastYear, &teamID,
			&p.TeamName, &p.DivisionID, &p.DivisionName, &p.League,
			&p.KPercent, &p.BBPercent, &p.InZonePercent,
			&p.WhiffPercent, &p.GroundballsPercent, &p.FlyballsPercent,
			&p.PopupsPercent, &p.PitchHand, &p.ArmAngle,
		)
		if err != nil {
			return nil, err
		}
		if birthDate.Valid {
			p.BirthDate = &birthDate.Time
		}
		if teamID.Valid {
			p.TeamID = int(teamID.Int64)
		}
		players = append(players, p)
	}
	return players, nil
}

// GetGuessesBySessionAndPuzzle fetches all guesses for a given session and puzzle.
func (r *Repository) GetGuessesBySessionAndPuzzle(sessionID string, puzzleID int) ([]models.Guess, error) {
	query := `
		SELECT id, session_id, puzzle_id, guessed_player_id, balls, strikes, result, created_at
		FROM guesses
		WHERE session_id = $1 AND puzzle_id = $2
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, sessionID, puzzleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guesses []models.Guess
	for rows.Next() {
		var g models.Guess
		err := rows.Scan(&g.ID, &g.SessionID, &g.PuzzleID, &g.GuessedPlayerID, &g.Balls, &g.Strikes, &g.Result, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		guesses = append(guesses, g)
	}
	return guesses, nil
}

// HasPlayerBeenGuessed checks if a player was already guessed in this session for this puzzle.
func (r *Repository) HasPlayerBeenGuessed(sessionID string, puzzleID int, playerID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM guesses WHERE session_id = $1 AND puzzle_id = $2 AND guessed_player_id = $3)",
		sessionID, puzzleID, playerID,
	).Scan(&exists)
	return exists, err
}

// SaveGuess saves a guess record.
func (r *Repository) SaveGuess(g *models.Guess) error {
	query := `
		INSERT INTO guesses (session_id, puzzle_id, guessed_player_id, balls, strikes, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, g.SessionID, g.PuzzleID, g.GuessedPlayerID, g.Balls, g.Strikes, g.Result).Scan(&g.ID, &g.CreatedAt)
}

// GetPitchGuessBySessionAndPuzzle fetches the pitch guess for a session and puzzle if it exists.
func (r *Repository) GetPitchGuessBySessionAndPuzzle(sessionID string, puzzleID int) (*models.PitchGuess, error) {
	query := `
		SELECT id, session_id, puzzle_id, guessed_pitch_type, matched, created_at
		FROM pitch_guesses
		WHERE session_id = $1 AND puzzle_id = $2
	`
	var pg models.PitchGuess
	err := r.db.QueryRow(query, sessionID, puzzleID).Scan(
		&pg.ID, &pg.SessionID, &pg.PuzzleID, &pg.GuessedPitchType, &pg.Matched, &pg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pg, nil
}

// SavePitchGuess saves a pitch guess record.
func (r *Repository) SavePitchGuess(g *models.PitchGuess) error {
	query := `
		INSERT INTO pitch_guesses (session_id, puzzle_id, guessed_pitch_type, matched, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, g.SessionID, g.PuzzleID, g.GuessedPitchType, g.Matched).Scan(&g.ID, &g.CreatedAt)
}

// GetOrCreateDailyPuzzle checks if a puzzle exists for the date. If not, it creates it.
func (r *Repository) GetOrCreateDailyPuzzle(date time.Time) (*models.DailyPuzzle, error) {
	truncatedDate := date.UTC().Truncate(24 * time.Hour)

	// Try fetching existing puzzle
	var dp models.DailyPuzzle
	query := `
		SELECT id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
		FROM daily_puzzles
		WHERE puzzle_date = $1
	`
	err := r.db.QueryRow(query, truncatedDate).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)
	if err == nil {
		return &dp, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check daily puzzle: %w", err)
	}

	// No puzzle found, let's create one deterministically.
	// 1. Fetch all eligible players (those with a pitch profile) sorted by ID
	playerIDs, profileIDs, err := r.getEligiblePitchers()
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible pitchers: %w", err)
	}
	if len(playerIDs) == 0 {
		return nil, fmt.Errorf("no eligible pitchers found in database")
	}

	// 2. Select pitcher deterministically
	daysSinceEpoch := int(truncatedDate.Unix() / 86400)
	idx := daysSinceEpoch % len(playerIDs)
	targetPlayerID := playerIDs[idx]
	targetPitchProfileID := profileIDs[idx]

	// 3. Insert new daily puzzle. We use ON CONFLICT DO NOTHING (or fetch on conflict) to prevent race conditions.
	insertQuery := `
		INSERT INTO daily_puzzles (puzzle_date, target_player_id, target_pitch_profile_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (puzzle_date) DO UPDATE SET puzzle_date = EXCLUDED.puzzle_date -- simple dummy update to force RETURNING if needed
		RETURNING id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
	`
	err = r.db.QueryRow(insertQuery, truncatedDate, targetPlayerID, targetPitchProfileID).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert daily puzzle: %w", err)
	}

	return &dp, nil
}

func (r *Repository) getEligiblePitchers() ([]int, []int, error) {
	// Select unique player IDs that have a pitch profile, prioritizing Statcast players with highest usage pitch
	query := `
		SELECT DISTINCT ON (p.id) p.id, pp.id
		FROM players p
		JOIN pitch_profiles pp ON pp.player_id = p.id
		WHERE p.k_percent IS NOT NULL AND p.k_percent > 0
		ORDER BY p.id ASC, pp.usage_percent DESC, pp.velocity DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var playerIDs []int
	var profileIDs []int
	for rows.Next() {
		var playerID, profileID int
		if err := rows.Scan(&playerID, &profileID); err != nil {
			return nil, nil, err
		}
		playerIDs = append(playerIDs, playerID)
		profileIDs = append(profileIDs, profileID)
	}
	return playerIDs, profileIDs, nil
}


// ResetDailyPuzzleForTest rotates today's daily puzzle to a different eligible pitcher/pitch profile
// and clears any guesses and pitch guesses for that puzzle.
func (r *Repository) ResetDailyPuzzleForTest(sessionID string) (*models.DailyPuzzle, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	playerIDs, profileIDs, err := r.getEligiblePitchers()
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible pitchers: %w", err)
	}
	if len(playerIDs) == 0 {
		return nil, fmt.Errorf("no eligible pitchers found in database")
	}

	// Check existing puzzle
	var dp models.DailyPuzzle
	query := `
		SELECT id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
		FROM daily_puzzles
		WHERE puzzle_date = $1
	`
	err = r.db.QueryRow(query, today).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)

	nextIdx := 0
	if err == nil {
		// Find current index and rotate to the next pitcher
		currentIdx := -1
		for i, pid := range playerIDs {
			if pid == dp.TargetPlayerID {
				currentIdx = i
				break
			}
		}
		if len(playerIDs) > 1 {
			if currentIdx >= 0 {
				nextIdx = (currentIdx + 1) % len(playerIDs)
			} else {
				nextIdx = 1
			}
		}
		targetPlayerID := playerIDs[nextIdx]
		targetPitchProfileID := profileIDs[nextIdx]

		// Delete guesses, pitch guesses, game completions, and cached animation for this puzzle
		_, _ = r.db.Exec("DELETE FROM guesses WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM pitch_guesses WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM game_completions WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM animations WHERE pitch_profile_id = $1", dp.TargetPitchProfileID)
		// Update daily puzzle
		updateQuery := `
			UPDATE daily_puzzles
			SET target_player_id = $1, target_pitch_profile_id = $2
			WHERE id = $3
			RETURNING id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
		`
		err = r.db.QueryRow(updateQuery, targetPlayerID, targetPitchProfileID, dp.ID).Scan(
			&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update daily puzzle: %w", err)
		}
		return &dp, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query daily puzzle: %w", err)
	}

	// Insert new daily puzzle if not exists
	targetPlayerID := playerIDs[0]
	targetPitchProfileID := profileIDs[0]
	insertQuery := `
		INSERT INTO daily_puzzles (puzzle_date, target_player_id, target_pitch_profile_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
	`
	err = r.db.QueryRow(insertQuery, today, targetPlayerID, targetPitchProfileID).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create daily puzzle: %w", err)
	}

	return &dp, nil
}
// SetTargetPitcherForTest sets today's daily puzzle to a specific pitcher by player ID.
func (r *Repository) SetTargetPitcherForTest(playerID int, sessionID string) (*models.DailyPuzzle, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Fetch a primary pitch profile for this player
	var pitchProfileID int
	err := r.db.QueryRow(`
		SELECT id FROM pitch_profiles 
		WHERE player_id = $1 
		ORDER BY usage_percent DESC, velocity DESC 
		LIMIT 1
	`, playerID).Scan(&pitchProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pitch profile for player %d: %w", playerID, err)
	}

	var dp models.DailyPuzzle
	query := `
		SELECT id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
		FROM daily_puzzles
		WHERE puzzle_date = $1
	`
	err = r.db.QueryRow(query, today).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)

	if err == nil {
		// Delete guesses, pitch guesses, game completions, and cached animation for this puzzle
		_, _ = r.db.Exec("DELETE FROM guesses WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM pitch_guesses WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM game_completions WHERE puzzle_id = $1", dp.ID)
		_, _ = r.db.Exec("DELETE FROM animations WHERE pitch_profile_id = $1", dp.TargetPitchProfileID)
		// Update daily puzzle
		updateQuery := `
			UPDATE daily_puzzles
			SET target_player_id = $1, target_pitch_profile_id = $2
			WHERE id = $3
			RETURNING id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
		`
		err = r.db.QueryRow(updateQuery, playerID, pitchProfileID, dp.ID).Scan(
			&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update daily puzzle: %w", err)
		}
		return &dp, nil
	}

	insertQuery := `
		INSERT INTO daily_puzzles (puzzle_date, target_player_id, target_pitch_profile_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, puzzle_date, target_player_id, target_pitch_profile_id, created_at
	`
	err = r.db.QueryRow(insertQuery, today, playerID, pitchProfileID).Scan(
		&dp.ID, &dp.PuzzleDate, &dp.TargetPlayerID, &dp.TargetPitchProfileID, &dp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create daily puzzle: %w", err)
	}

	return &dp, nil
}

func formatPlayerName(sessionID string) string {
	cleanID := strings.TrimSpace(sessionID)
	if cleanID == "" || strings.EqualFold(cleanID, "anonymous") {
		return "Player-Anon"
	}
	cleanID = strings.TrimPrefix(cleanID, "session-")
	if len(cleanID) > 6 {
		cleanID = cleanID[:6]
	}
	return "Player-" + strings.ToUpper(cleanID)
}

// RecordGameCompletion records game result and updates user streak statistics.
func (r *Repository) RecordGameCompletion(sessionID string, puzzleID int, status string, guessCount int, pitchMatched bool, timeTakenSec int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var puzzleDate time.Time
	err = tx.QueryRow("SELECT puzzle_date FROM daily_puzzles WHERE id = $1", puzzleID).Scan(&puzzleDate)
	if err != nil {
		return fmt.Errorf("failed to find puzzle: %w", err)
	}
	puzzleDate = puzzleDate.UTC().Truncate(24 * time.Hour)

	// Check if already completed
	var existingID int
	var prevStatus string
	err = tx.QueryRow("SELECT id, status FROM game_completions WHERE session_id = $1 AND puzzle_id = $2", sessionID, puzzleID).Scan(&existingID, &prevStatus)
	alreadyCompleted := (err == nil)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing completion: %w", err)
	}

	compQuery := `
		INSERT INTO game_completions (session_id, puzzle_id, status, guess_count, pitch_matched, time_taken_seconds, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (session_id, puzzle_id) DO UPDATE SET
			status = EXCLUDED.status,
			guess_count = EXCLUDED.guess_count,
			pitch_matched = EXCLUDED.pitch_matched,
			time_taken_seconds = EXCLUDED.time_taken_seconds,
			completed_at = NOW()
	`
	_, err = tx.Exec(compQuery, sessionID, puzzleID, status, guessCount, pitchMatched, timeTakenSec)
	if err != nil {
		return fmt.Errorf("failed to record game completion: %w", err)
	}

	// Only update user streak stats if not already completed for this puzzle
	if !alreadyCompleted && sessionID != "" && !strings.EqualFold(sessionID, "anonymous") {
		var streak models.UserStreak
		var lastWon sql.NullTime
		streakQuery := `
			SELECT session_id, games_played, games_won, current_streak, max_streak, last_won_puzzle_date
			FROM user_streaks
			WHERE session_id = $1
		`
		err = tx.QueryRow(streakQuery, sessionID).Scan(
			&streak.SessionID, &streak.GamesPlayed, &streak.GamesWon,
			&streak.CurrentStreak, &streak.MaxStreak, &lastWon,
		)

		if err == sql.ErrNoRows {
			streak.SessionID = sessionID
			streak.GamesPlayed = 1
			if status == "won" {
				streak.GamesWon = 1
				streak.CurrentStreak = 1
				streak.MaxStreak = 1
				_, err = tx.Exec(`
					INSERT INTO user_streaks (session_id, games_played, games_won, current_streak, max_streak, last_won_puzzle_date, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, NOW())
				`, sessionID, streak.GamesPlayed, streak.GamesWon, streak.CurrentStreak, streak.MaxStreak, puzzleDate)
			} else {
				streak.GamesWon = 0
				streak.CurrentStreak = 0
				streak.MaxStreak = 0
				_, err = tx.Exec(`
					INSERT INTO user_streaks (session_id, games_played, games_won, current_streak, max_streak, last_won_puzzle_date, updated_at)
					VALUES ($1, $2, $3, $4, $5, NULL, NOW())
				`, sessionID, streak.GamesPlayed, streak.GamesWon, streak.CurrentStreak, streak.MaxStreak)
			}
			if err != nil {
				return fmt.Errorf("failed to insert user streak: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to fetch user streak: %w", err)
		} else {
			streak.GamesPlayed += 1
			if status == "won" {
				streak.GamesWon += 1
				if lastWon.Valid {
					prevDate := lastWon.Time.UTC().Truncate(24 * time.Hour)
					daysDiff := int(puzzleDate.Sub(prevDate).Hours() / 24)
					if daysDiff == 1 {
						streak.CurrentStreak += 1
					} else if daysDiff == 0 {
						if streak.CurrentStreak == 0 {
							streak.CurrentStreak = 1
						}
					} else {
						streak.CurrentStreak = 1
					}
				} else {
					streak.CurrentStreak = 1
				}

				if streak.CurrentStreak > streak.MaxStreak {
					streak.MaxStreak = streak.CurrentStreak
				}
				_, err = tx.Exec(`
					UPDATE user_streaks
					SET games_played = $1, games_won = $2, current_streak = $3, max_streak = $4, last_won_puzzle_date = $5, updated_at = NOW()
					WHERE session_id = $6
				`, streak.GamesPlayed, streak.GamesWon, streak.CurrentStreak, streak.MaxStreak, puzzleDate, sessionID)
			} else {
				streak.CurrentStreak = 0
				_, err = tx.Exec(`
					UPDATE user_streaks
					SET games_played = $1, games_won = $2, current_streak = 0, updated_at = NOW()
					WHERE session_id = $3
				`, streak.GamesPlayed, streak.GamesWon, sessionID)
			}
			if err != nil {
				return fmt.Errorf("failed to update user streak: %w", err)
			}
		}
	}

	return tx.Commit()
}

// GetTodayPuzzleStats calculates global puzzle stats and personal user streak stats.
func (r *Repository) GetTodayPuzzleStats(puzzleID int, sessionID string) (*models.DailyStats, error) {
	var totalSolved int
	err := r.db.QueryRow("SELECT COUNT(*) FROM game_completions WHERE puzzle_id = $1 AND status = 'won'", puzzleID).Scan(&totalSolved)
	if err != nil {
		return nil, fmt.Errorf("failed to count total solved: %w", err)
	}

	var totalAttempts int
	attemptQuery := `
		SELECT COUNT(DISTINCT s_id) FROM (
			SELECT session_id AS s_id FROM guesses WHERE puzzle_id = $1
			UNION
			SELECT session_id AS s_id FROM game_completions WHERE puzzle_id = $1
		) t
	`
	err = r.db.QueryRow(attemptQuery, puzzleID).Scan(&totalAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to count total attempts: %w", err)
	}
	if totalSolved > totalAttempts {
		totalAttempts = totalSolved
	}

	var winRate float64
	if totalAttempts > 0 {
		winRate = math.Round((float64(totalSolved)/float64(totalAttempts)*100.0)*10) / 10
	}

	distribution := models.GuessDistributionMap{
		"1": 0, "2": 0, "3": 0, "4": 0, "5": 0,
		"6": 0, "7": 0, "8": 0, "9+": 0,
	}

	distRows, err := r.db.Query("SELECT guess_count, COUNT(*) FROM game_completions WHERE puzzle_id = $1 AND status = 'won' GROUP BY guess_count", puzzleID)
	if err == nil {
		defer distRows.Close()
		for distRows.Next() {
			var gCount, count int
			if err := distRows.Scan(&gCount, &count); err == nil {
				if gCount >= 1 && gCount <= 8 {
					distribution[fmt.Sprintf("%d", gCount)] = count
				} else if gCount >= 9 {
					distribution["9+"] += count
				}
			}
		}
	}

	userStats := &models.UserStats{
		GamesPlayed:    0,
		GamesWon:       0,
		CurrentStreak:  0,
		MaxStreak:      0,
		UserGuessCount: 0,
		Solved:         false,
	}

	if sessionID != "" && !strings.EqualFold(sessionID, "anonymous") {
		var s models.UserStreak
		err = r.db.QueryRow("SELECT games_played, games_won, current_streak, max_streak FROM user_streaks WHERE session_id = $1", sessionID).Scan(
			&s.GamesPlayed, &s.GamesWon, &s.CurrentStreak, &s.MaxStreak,
		)
		if err == nil {
			userStats.GamesPlayed = s.GamesPlayed
			userStats.GamesWon = s.GamesWon
			userStats.CurrentStreak = s.CurrentStreak
			userStats.MaxStreak = s.MaxStreak
		}

		var comp models.GameCompletion
		err = r.db.QueryRow("SELECT guess_count, status FROM game_completions WHERE session_id = $1 AND puzzle_id = $2", sessionID, puzzleID).Scan(
			&comp.GuessCount, &comp.Status,
		)
		if err == nil {
			userStats.UserGuessCount = comp.GuessCount
			userStats.Solved = (comp.Status == "won")
		} else {
			var guessCount int
			_ = r.db.QueryRow("SELECT COUNT(*) FROM guesses WHERE session_id = $1 AND puzzle_id = $2", sessionID, puzzleID).Scan(&guessCount)
			userStats.UserGuessCount = guessCount
		}
	}

	return &models.DailyStats{
		TotalSolved:   totalSolved,
		TotalAttempts: totalAttempts,
		WinRate:       winRate,
		Distribution:  distribution,
		UserStats:     userStats,
	}, nil
}

// GetDailyLeaderboard returns top solvers for today's daily puzzle.
func (r *Repository) GetDailyLeaderboard(puzzleID int, limit int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		SELECT session_id, guess_count, pitch_matched, time_taken_seconds, completed_at
		FROM game_completions
		WHERE puzzle_id = $1 AND status = 'won'
		ORDER BY guess_count ASC, time_taken_seconds ASC, completed_at ASC
		LIMIT $2
	`
	rows, err := r.db.Query(query, puzzleID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]models.LeaderboardEntry, 0)
	rank := 1
	for rows.Next() {
		var sessionID string
		var guessCount int
		var pitchMatched bool
		var timeTakenSec int
		var completedAt time.Time

		err := rows.Scan(&sessionID, &guessCount, &pitchMatched, &timeTakenSec, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}

		entries = append(entries, models.LeaderboardEntry{
			Rank:         rank,
			PlayerName:   formatPlayerName(sessionID),
			GuessCount:   guessCount,
			PitchMatched: pitchMatched,
			TimeSeconds:  timeTakenSec,
			CompletedAt:  completedAt,
		})
		rank++
	}

	return entries, nil
}

// GetStreakLeaderboard returns top players by active win streak.
func (r *Repository) GetStreakLeaderboard(limit int) ([]models.StreakLeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		SELECT session_id, current_streak, max_streak, games_played, games_won
		FROM user_streaks
		WHERE current_streak > 0 OR games_played > 0
		ORDER BY current_streak DESC, max_streak DESC, games_won DESC, updated_at ASC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch streak leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]models.StreakLeaderboardEntry, 0)
	rank := 1
	for rows.Next() {
		var sessionID string
		var currentStreak, maxStreak, gamesPlayed, gamesWon int

		err := rows.Scan(&sessionID, &currentStreak, &maxStreak, &gamesPlayed, &gamesWon)
		if err != nil {
			return nil, fmt.Errorf("failed to scan streak leaderboard row: %w", err)
		}

		var winRate float64
		if gamesPlayed > 0 {
			winRate = math.Round((float64(gamesWon)/float64(gamesPlayed)*100.0)*10) / 10
		}

		entries = append(entries, models.StreakLeaderboardEntry{
			Rank:          rank,
			PlayerName:    formatPlayerName(sessionID),
			CurrentStreak: currentStreak,
			MaxStreak:     maxStreak,
			WinRate:       winRate,
		})
		rank++
	}

	return entries, nil
}