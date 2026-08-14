package repository

import (
	"database/sql"
	"fmt"
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
		       COALESCE(t.name, '') AS team_name, COALESCE(t.division_id, 0) AS division_id, COALESCE(d.name, '') AS division_name
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
		&p.TeamName, &p.DivisionID, &p.DivisionName,
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
		       release_extension, break_x, break_z, arm_angle, plate_x, plate_z, created_at
		FROM pitch_profiles
		WHERE id = $1
	`
	var p models.PitchProfile
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.PlayerID, &p.PitchType, &p.Velocity, &p.SpinRate, &p.ReleasePosX, &p.ReleasePosZ,
		&p.ReleaseExtension, &p.BreakX, &p.BreakZ, &p.ArmAngle, &p.PlateX, &p.PlateZ, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SearchPlayers searches for players by name (case-insensitive substring) returning up to 10 results.
func (r *Repository) SearchPlayers(query string) ([]models.Player, error) {
	sqlQuery := `
		SELECT p.id, p.mlb_id, p.name, p.birth_date, p.birth_year, p.position, p.mlb_debut_year, p.mlb_last_year, p.team_id,
		       COALESCE(t.name, '') AS team_name, COALESCE(t.division_id, 0) AS division_id, COALESCE(d.name, '') AS division_name
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
			&p.TeamName, &p.DivisionID, &p.DivisionName,
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
	// Select player IDs that have a pitch profile, ordered by player id
	query := `
		SELECT p.id, pp.id
		FROM players p
		JOIN pitch_profiles pp ON pp.player_id = p.id
		ORDER BY p.id ASC
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
