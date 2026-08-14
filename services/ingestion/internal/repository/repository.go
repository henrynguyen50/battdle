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

// UpsertPlayerWithProfiles upserts a player and all their pitch profiles in a transaction
func (r *Repository) UpsertPlayerWithProfiles(player models.Player, name string, profiles []models.PitchProfile) error {
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
			position, height, weight, mlb_debut_year, mlb_last_year, team_id,
			k_percent, bb_percent, in_zone_percent, whiff_percent,
			groundballs_percent, flyballs_percent, popups_percent, pitch_hand, arm_angle,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW())
		ON CONFLICT (mlb_id) DO UPDATE SET
			name = EXCLUDED.name,
			birth_date = COALESCE(EXCLUDED.birth_date, players.birth_date),
			birth_year = EXCLUDED.birth_year,
			birth_city = COALESCE(EXCLUDED.birth_city, players.birth_city),
			birth_country = COALESCE(EXCLUDED.birth_country, players.birth_country),
			position = EXCLUDED.position,
			height = COALESCE(EXCLUDED.height, players.height),
			weight = COALESCE(EXCLUDED.weight, players.weight),
			mlb_debut_year = EXCLUDED.mlb_debut_year,
			mlb_last_year = EXCLUDED.mlb_last_year,
			team_id = EXCLUDED.team_id,
			k_percent = EXCLUDED.k_percent,
			bb_percent = EXCLUDED.bb_percent,
			in_zone_percent = EXCLUDED.in_zone_percent,
			whiff_percent = EXCLUDED.whiff_percent,
			groundballs_percent = EXCLUDED.groundballs_percent,
			flyballs_percent = EXCLUDED.flyballs_percent,
			popups_percent = EXCLUDED.popups_percent,
			pitch_hand = EXCLUDED.pitch_hand,
			arm_angle = EXCLUDED.arm_angle,
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
		player.KPercent,
		player.BBPercent,
		player.InZonePercent,
		player.WhiffPercent,
		player.GroundballsPercent,
		player.FlyballsPercent,
		player.PopupsPercent,
		player.PitchHand,
		player.ArmAngle,
	).Scan(&playerID)
	if err != nil {
		return fmt.Errorf("failed to upsert player %s: %w", name, err)
	}

	// 2. Delete existing pitch profile(s) for this player
	_, err = tx.Exec("DELETE FROM pitch_profiles WHERE player_id = $1", playerID)
	if err != nil {
		return fmt.Errorf("failed to delete existing pitch profiles: %w", err)
	}

	// 3. Insert new pitch profiles
	queryProfile := `
		INSERT INTO pitch_profiles (
			player_id, pitch_type, velocity, spin_rate, release_pos_x, release_pos_z,
			release_extension, break_x, break_z, arm_angle, plate_x, plate_z,
			usage_percent, break_z_induced, range_speed
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	for _, profile := range profiles {
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
			profile.UsagePercent,
			profile.BreakZInduced,
			profile.RangeSpeed,
		)
		if err != nil {
			return fmt.Errorf("failed to insert pitch profile (%s): %w", profile.PitchType, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpsertPlayerAndProfile upserts a player and replaces their single pitch profile
func (r *Repository) UpsertPlayerAndProfile(player models.Player, name string, profile models.PitchProfile) error {
	return r.UpsertPlayerWithProfiles(player, name, []models.PitchProfile{profile})
}
