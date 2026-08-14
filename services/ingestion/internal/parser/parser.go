package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type PitchRecord struct {
	PlayerID         int
	PlayerName       string
	SpinRate         float64
	Velocity         float64
	ReleaseExtension float64
	BreakX           float64
	BreakZ           float64
	ReleasePosX      float64
	ReleasePosZ      float64
	PlateX           float64
	PlateZ           float64
	ArmAngle         float64
}

func ParseCSV(filePath string) ([]PitchRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map headers to indices
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[name] = i
	}

	requiredCols := []string{
		"player_id", "player_name", "spin_rate", "velocity",
		"release_extension", "api_break_x_arm", "api_break_z_with_gravity",
		"release_pos_x", "release_pos_z", "plate_x", "plate_z", "arm_angle",
	}

	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			return nil, fmt.Errorf("missing required CSV column: %s", col)
		}
	}

	var records []PitchRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		playerID, err := strconv.Atoi(row[colMap["player_id"]])
		if err != nil {
			continue // Skip invalid rows
		}

		playerName := row[colMap["player_name"]]

		spinRate, _ := strconv.ParseFloat(row[colMap["spin_rate"]], 64)
		velocity, _ := strconv.ParseFloat(row[colMap["velocity"]], 64)
		releaseExt, _ := strconv.ParseFloat(row[colMap["release_extension"]], 64)
		breakX, _ := strconv.ParseFloat(row[colMap["api_break_x_arm"]], 64)
		breakZ, _ := strconv.ParseFloat(row[colMap["api_break_z_with_gravity"]], 64)
		releaseX, _ := strconv.ParseFloat(row[colMap["release_pos_x"]], 64)
		releaseZ, _ := strconv.ParseFloat(row[colMap["release_pos_z"]], 64)
		plateX, _ := strconv.ParseFloat(row[colMap["plate_x"]], 64)
		plateZ, _ := strconv.ParseFloat(row[colMap["plate_z"]], 64)
		armAngle, _ := strconv.ParseFloat(row[colMap["arm_angle"]], 64)

		records = append(records, PitchRecord{
			PlayerID:         playerID,
			PlayerName:       playerName,
			SpinRate:         spinRate,
			Velocity:         velocity,
			ReleaseExtension: releaseExt,
			BreakX:           breakX,
			BreakZ:           breakZ,
			ReleasePosX:      releaseX,
			ReleasePosZ:      releaseZ,
			PlateX:           plateX,
			PlateZ:           plateZ,
			ArmAngle:         armAngle,
		})
	}

	return records, nil
}

type MetadataRecord struct {
	PlayerID     int
	PlayerName   string
	BirthDate    string
	BirthYear    int
	BirthCity    string
	BirthCountry string
	Position     string
	Height       string
	Weight       int
	MLBDebutYear int
	MLBLastYear  int
	MLBTeamID    int
}

func ParseMetadataCSV(filePath string) (map[int]MetadataRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata CSV header: %w", err)
	}

	// Map headers to indices
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[name] = i
	}

	requiredCols := []string{
		"player_id", "player_name", "birth_date", "birth_year",
		"birth_city", "birth_country", "position", "height", "weight",
		"mlb_debut_year", "mlb_last_year", "mlb_team_id",
	}

	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			return nil, fmt.Errorf("missing required metadata CSV column: %s", col)
		}
	}

	records := make(map[int]MetadataRecord)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading metadata CSV row: %w", err)
		}

		playerID, err := strconv.Atoi(row[colMap["player_id"]])
		if err != nil {
			continue // Skip invalid rows
		}

		birthYear, _ := strconv.Atoi(row[colMap["birth_year"]])
		weight, _ := strconv.Atoi(row[colMap["weight"]])
		mlbDebutYear, _ := strconv.Atoi(row[colMap["mlb_debut_year"]])
		mlbLastYear, _ := strconv.Atoi(row[colMap["mlb_last_year"]])
		mlbTeamID, _ := strconv.Atoi(row[colMap["mlb_team_id"]])

		records[playerID] = MetadataRecord{
			PlayerID:     playerID,
			PlayerName:   row[colMap["player_name"]],
			BirthDate:    row[colMap["birth_date"]],
			BirthYear:    birthYear,
			BirthCity:    row[colMap["birth_city"]],
			BirthCountry: row[colMap["birth_country"]],
			Position:     row[colMap["position"]],
			Height:       row[colMap["height"]],
			Weight:       weight,
			MLBDebutYear: mlbDebutYear,
			MLBLastYear:  mlbLastYear,
			MLBTeamID:    mlbTeamID,
		}
	}

	return records, nil
}
