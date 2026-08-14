package models

import (
	"time"
)

type Division struct {
	ID     int    `json:"id" db:"id"`
	League string `json:"league" db:"league"`
	Name   string `json:"name" db:"name"`
}

type Team struct {
	ID         int    `json:"id" db:"id"`
	MLBTeamID  int    `json:"mlb_team_id" db:"mlb_team_id"`
	Name       string `json:"name" db:"name"`
	DivisionID int    `json:"division_id" db:"division_id"`
}

type Player struct {
	ID           int        `json:"id" db:"id"`
	MLBID        int        `json:"mlb_id" db:"mlb_id"`
	Name         string     `json:"name" db:"name"`
	BirthDate    *time.Time `json:"birth_date" db:"birth_date"`
	BirthYear    int        `json:"birth_year" db:"birth_year"`
	BirthCity    string     `json:"birth_city,omitempty" db:"birth_city"`
	BirthCountry string     `json:"birth_country,omitempty" db:"birth_country"`
	Position     string     `json:"position" db:"position"`
	Height       string     `json:"height,omitempty" db:"height"`
	Weight       int        `json:"weight,omitempty" db:"weight"`
	MLBDebutYear int        `json:"mlb_debut_year" db:"mlb_debut_year"`
	MLBLastYear  int        `json:"mlb_last_year" db:"mlb_last_year"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	DivisionID   int    `json:"division_id,omitempty" db:"division_id"`
	DivisionName string `json:"division_name,omitempty" db:"division_name"`
	TeamID       int    `json:"team_id,omitempty" db:"team_id"`
	TeamName     string `json:"team_name,omitempty" db:"team_name"`
}

type PitchProfile struct {
	ID               int       `json:"id" db:"id"`
	PlayerID         int       `json:"player_id" db:"player_id"`
	PitchType        string    `json:"pitch_type" db:"pitch_type"`
	Velocity         float64   `json:"velocity" db:"velocity"`
	SpinRate         float64   `json:"spin_rate" db:"spin_rate"`
	ReleasePosX      float64   `json:"release_pos_x" db:"release_pos_x"`
	ReleasePosZ      float64   `json:"release_pos_z" db:"release_pos_z"`
	ReleaseExtension float64   `json:"release_extension" db:"release_extension"`
	BreakX           float64   `json:"break_x" db:"break_x"`
	BreakZ           float64   `json:"break_z" db:"break_z"`
	ArmAngle         float64   `json:"arm_angle" db:"arm_angle"`
	PlateX           float64   `json:"plate_x" db:"plate_x"`
	PlateZ           float64   `json:"plate_z" db:"plate_z"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type DailyPuzzle struct {
	ID                   int       `json:"id" db:"id"`
	PuzzleDate           time.Time `json:"puzzle_date" db:"puzzle_date"`
	TargetPlayerID       int       `json:"target_player_id" db:"target_player_id"`
	TargetPitchProfileID int       `json:"target_pitch_profile_id" db:"target_pitch_profile_id"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

type Guess struct {
	ID              int       `json:"id" db:"id"`
	SessionID       string    `json:"session_id" db:"session_id"`
	PuzzleID        int       `json:"puzzle_id" db:"puzzle_id"`
	GuessedPlayerID int       `json:"guessed_player_id" db:"guessed_player_id"`
	Balls           int       `json:"balls" db:"balls"`
	Strikes         int       `json:"strikes" db:"strikes"`
	Result          string    `json:"result" db:"result"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Animation struct {
	ID             int       `json:"id" db:"id"`
	PitchProfileID int       `json:"pitch_profile_id" db:"pitch_profile_id"`
	AnimationData  string    `json:"animation_data" db:"animation_data"` // JSONB stored as string
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
