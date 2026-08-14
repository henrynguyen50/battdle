package repository

import (
	"database/sql"
	"fmt"

	"pitchle/shared/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetTeamMap returns a map of mlb_team_id to DB id
func (r *Repository) GetTeamMap() (map[int]int, error) {
	rows, err := r.db.Query("SELECT id, mlb_team_id FROM teams")
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	teamMap := make(map[int]int)
	for rows.Next() {
		var id, mlbTeamID int
		if err := rows.Scan(&id, &mlbTeamID); err != nil {
			return nil, fmt.Errorf("failed to scan team row: %w", err)
		}
		teamMap[mlbTeamID] = id
	}

	return teamMap, nil
}

// UpsertPlayerAndProfile upserts a player and replaces their pitch profile in a transaction
func (r *Repository) UpsertPlayerAndProfile(player models.Player, name string, profile models.PitchProfile) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Upsert player
	var playerID int
	queryPlayer := `
		INSERT INTO players (
			mlb_id, name, birth_date, birth_year, birth_city, birth_country,
			position, height, weight, mlb_debut_year, mlb_last_year, team_id, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (mlb_id) DO UPDATE SET
			name = EXCLUDED.name,
			birth_date = EXCLUDED.birth_date,
			birth_year = EXCLUDED.birth_year,
			birth_city = EXCLUDED.birth_city,
			birth_country = EXCLUDED.birth_country,
			position = EXCLUDED.position,
			height = EXCLUDED.height,
			weight = EXCLUDED.weight,
			mlb_debut_year = EXCLUDED.mlb_debut_year,
			mlb_last_year = EXCLUDED.mlb_last_year,
			team_id = EXCLUDED.team_id,
			updated_at = NOW()
		RETURNING id
	`
	err = tx.QueryRow(
		queryPlayer,
		player.MLBID,
		name,
		player.BirthDate,
		player.BirthYear,
		player.BirthCity,
		player.BirthCountry,
		player.Position,
		player.Height,
		player.Weight,
		player.MLBDebutYear,
		player.MLBLastYear,
		player.TeamID,
	).Scan(&playerID)
	if err != nil {
		return fmt.Errorf("failed to upsert player %s: %w", name, err)
	}

	// 2. Delete existing pitch profile(s) for this player
	_, err = tx.Exec("DELETE FROM pitch_profiles WHERE player_id = $1", playerID)
	if err != nil {
		return fmt.Errorf("failed to delete existing pitch profiles: %w", err)
	}

	// 3. Insert new pitch profile
	queryProfile := `
		INSERT INTO pitch_profiles (
			player_id, pitch_type, velocity, spin_rate, release_pos_x, release_pos_z,
			release_extension, break_x, break_z, arm_angle, plate_x, plate_z
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.Exec(
		queryProfile,
		playerID,
		profile.PitchType,
		profile.Velocity,
		profile.SpinRate,
		profile.ReleasePosX,
		profile.ReleasePosZ,
		profile.ReleaseExtension,
		profile.BreakX,
		profile.BreakZ,
		profile.ArmAngle,
		profile.PlateX,
		profile.PlateZ,
	)
	if err != nil {
		return fmt.Errorf("failed to insert pitch profile: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
