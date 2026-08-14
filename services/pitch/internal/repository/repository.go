package repository

import (
	"database/sql"
	"time"

	"pitchle/shared/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetTodayPitchProfileID fetches the target_pitch_profile_id for today's daily puzzle.
func (r *Repository) GetTodayPitchProfileID(date time.Time) (int, error) {
	truncatedDate := date.UTC().Truncate(24 * time.Hour)
	var profileID int
	query := "SELECT target_pitch_profile_id FROM daily_puzzles WHERE puzzle_date = $1"
	err := r.db.QueryRow(query, truncatedDate).Scan(&profileID)
	return profileID, err
}

// GetPitchProfileByID fetches a pitch profile by ID.
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

// GetAnimationByProfileID fetches animation data for a pitch profile.
func (r *Repository) GetAnimationByProfileID(profileID int) (string, error) {
	var animData string
	query := "SELECT animation_data FROM animations WHERE pitch_profile_id = $1"
	err := r.db.QueryRow(query, profileID).Scan(&animData)
	return animData, err
}

// SaveAnimation stores animation data for a pitch profile.
func (r *Repository) SaveAnimation(profileID int, animationData string) error {
	query := `
		INSERT INTO animations (pitch_profile_id, animation_data, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (pitch_profile_id) DO UPDATE SET animation_data = EXCLUDED.animation_data, created_at = NOW()
	`
	_, err := r.db.Exec(query, profileID, animationData)
	return err
}
