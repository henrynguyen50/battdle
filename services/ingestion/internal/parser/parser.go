package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)
func stripBOM(r io.Reader) io.Reader {
	br := bufio.NewReader(r)
	r3, err := br.Peek(3)
	if err == nil && len(r3) >= 3 && r3[0] == 0xef && r3[1] == 0xbb && r3[2] == 0xbf {
		br.Discard(3)
	}
	return br
}

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

	reader := csv.NewReader(stripBOM(file))
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

	reader := csv.NewReader(stripBOM(file))
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

type PitchData struct {
	PitchType        string
	UsagePercent     float64
	Velocity         float64
	SpinRate         float64
	BreakX           float64
	BreakZ           float64
	BreakZInduced    float64
	RangeSpeed       float64
	ReleasePosX      float64
	ReleasePosZ      float64
	ReleaseExtension float64
	PlateX           float64
	PlateZ           float64
	ArmAngle         float64
}

type PlayerStatcastRecord struct {
	PlayerID           int
	RawName            string
	NormalizedName     string
	Year               int
	PlayerAge          int
	KPercent           float64
	BBPercent          float64
	InZonePercent      float64
	WhiffPercent       float64
	GroundballsPercent float64
	FlyballsPercent    float64
	PopupsPercent      float64
	PitchHand          string
	ArmAngle           float64
	Pitches            []PitchData
}

var PitchPrefixes = []struct {
	Prefix string
	Name   string
}{
	{"ff", "Four-Seam Fastball"},
	{"si", "Sinker"},
	{"fc", "Cutter"},
	{"sl", "Slider"},
	{"st", "Sweeper"},
	{"cu", "Curveball"},
	{"ch", "Changeup"},
	{"fs", "Splitter"},
	{"kn", "Knuckleball"},
	{"sv", "Slurve"},
	{"fo", "Forkball"},
	{"sc", "Screwball"},
}

func ParseUpdatedCSV(filePath string) ([]PlayerStatcastRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open updated CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(stripBOM(file))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read updated CSV header: %w", err)
	}

	colMap := make(map[string]int)
	for i, name := range header {
		cleaned := strings.TrimSpace(name)
		cleaned = strings.Trim(cleaned, "\"")
		colMap[cleaned] = i
	}

	// Locate name column
	nameCol := -1
	for _, candidate := range []string{"last_name, first_name", "player_name", "name"} {
		if idx, ok := colMap[candidate]; ok {
			nameCol = idx
			break
		}
	}
	if nameCol == -1 {
		// Check if first column contains name
		if len(header) > 0 {
			nameCol = 0
		}
	}

	pidCol, ok := colMap["player_id"]
	if !ok {
		return nil, fmt.Errorf("missing player_id column in updated CSV")
	}

	var records []PlayerStatcastRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading updated CSV row: %w", err)
		}

		if len(row) <= pidCol {
			continue
		}

		playerID, err := strconv.Atoi(strings.TrimSpace(row[pidCol]))
		if err != nil || playerID == 0 {
			continue
		}

		rawName := ""
		if nameCol >= 0 && nameCol < len(row) {
			rawName = strings.TrimSpace(row[nameCol])
		}

		// Normalize "Last, First" -> "First Last"
		normalizedName := rawName
		parts := strings.Split(rawName, ",")
		if len(parts) == 2 {
			normalizedName = strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
		}

		yearVal := 2026
		if idx, ok := colMap["year"]; ok && idx < len(row) {
			if y, err := strconv.Atoi(strings.TrimSpace(row[idx])); err == nil && y > 0 {
				yearVal = y
			}
		}

		var playerAge int
		if idx, ok := colMap["player_age"]; ok && idx < len(row) {
			playerAge, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
		}

		var kPct, bbPct, inZonePct, whiffPct, gbPct, fbPct, popupPct, armAngle float64
		if idx, ok := colMap["k_percent"]; ok && idx < len(row) {
			kPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["bb_percent"]; ok && idx < len(row) {
			bbPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["in_zone_percent"]; ok && idx < len(row) {
			inZonePct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["whiff_percent"]; ok && idx < len(row) {
			whiffPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["groundballs_percent"]; ok && idx < len(row) {
			gbPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["flyballs_percent"]; ok && idx < len(row) {
			fbPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["popups_percent"]; ok && idx < len(row) {
			popupPct, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}
		if idx, ok := colMap["arm_angle"]; ok && idx < len(row) {
			armAngle, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		}

		pitchHand := "R"
		if idx, ok := colMap["pitch_hand"]; ok && idx < len(row) {
			h := strings.TrimSpace(row[idx])
			if h == "L" || h == "l" {
				pitchHand = "L"
			}
		}

		// Release position: RHP releases from screen right (+X), LHP from screen left (-X)
		armRad := (armAngle * 3.1415926535) / 180.0
		if armRad <= 0 {
			armRad = 0.65
		}
		var relPosX float64
		if pitchHand == "L" {
			relPosX = +(1.2 + 1.4*math.Cos(armRad)) // LHP releases on first-base side (+X)
		} else {
			relPosX = -(1.2 + 1.4*math.Cos(armRad)) // RHP releases on third-base side (-X)
		}
		relPosZ := 4.8 + 1.5*math.Sin(armRad)
		relExt := 6.2
		// Extract individual pitches
		var pitches []PitchData
		for _, pp := range PitchPrefixes {
			usageIdx, hasUsage := colMap[fmt.Sprintf("n_%s_formatted", pp.Prefix)]
			speedIdx, hasSpeed := colMap[fmt.Sprintf("%s_avg_speed", pp.Prefix)]

			if !hasUsage || !hasSpeed || usageIdx >= len(row) || speedIdx >= len(row) {
				continue
			}

			usageStr := strings.TrimSpace(row[usageIdx])
			speedStr := strings.TrimSpace(row[speedIdx])
			if usageStr == "" || speedStr == "" {
				continue
			}

			usage, err1 := strconv.ParseFloat(usageStr, 64)
			speed, err2 := strconv.ParseFloat(speedStr, 64)
			if err1 != nil || err2 != nil || usage <= 0 || speed <= 0 {
				continue
			}

			var spin, breakX, breakZ, breakZInd, rangeSpeed float64
			if idx, ok := colMap[fmt.Sprintf("%s_avg_spin", pp.Prefix)]; ok && idx < len(row) {
				spin, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			}
			if idx, ok := colMap[fmt.Sprintf("%s_avg_break_x", pp.Prefix)]; ok && idx < len(row) {
				breakX, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			}
			if idx, ok := colMap[fmt.Sprintf("%s_avg_break_z", pp.Prefix)]; ok && idx < len(row) {
				breakZ, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			}
			if idx, ok := colMap[fmt.Sprintf("%s_avg_break_z_induced", pp.Prefix)]; ok && idx < len(row) {
				breakZInd, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			}
			if idx, ok := colMap[fmt.Sprintf("%s_range_speed", pp.Prefix)]; ok && idx < len(row) {
				rangeSpeed, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			}

			plateX, plateZ := computeRealisticPlateLocation(pp.Name, pitchHand, playerID)

			pitches = append(pitches, PitchData{
				PitchType:        pp.Name,
				UsagePercent:     usage,
				Velocity:         speed,
				SpinRate:         spin,
				BreakX:           breakX,
				BreakZ:           breakZ,
				BreakZInduced:    breakZInd,
				RangeSpeed:       rangeSpeed,
				ReleasePosX:      relPosX,
				ReleasePosZ:      relPosZ,
				ReleaseExtension: relExt,
				PlateX:           plateX,
				PlateZ:           plateZ,
				ArmAngle:         armAngle,
			})
		}

		// Sort pitches by usage descending
		sort.Slice(pitches, func(i, j int) bool {
			return pitches[i].UsagePercent > pitches[j].UsagePercent
		})

		records = append(records, PlayerStatcastRecord{
			PlayerID:           playerID,
			RawName:            rawName,
			NormalizedName:     normalizedName,
			Year:               yearVal,
			PlayerAge:          playerAge,
			KPercent:           kPct,
			BBPercent:          bbPct,
			InZonePercent:      inZonePct,
			WhiffPercent:       whiffPct,
			GroundballsPercent: gbPct,
			FlyballsPercent:    fbPct,
			PopupsPercent:      popupPct,
			PitchHand:          pitchHand,
			ArmAngle:           armAngle,
			Pitches:            pitches,
		})
	}

	return records, nil
}

func computeRealisticPlateLocation(pitchType string, pitchHand string, playerID int) (float64, float64) {
	isRHP := pitchHand != "L" && pitchHand != "l"
	armFactor := 1.0
	if !isRHP {
		armFactor = -1.0
	}

	// Deterministic subtle variance per player (±0.06 ft)
	jitterX := (float64((playerID*17)%100) / 500.0) - 0.1
	jitterZ := (float64((playerID*31)%100) / 500.0) - 0.1

	var plateX, plateZ float64
	switch pitchType {
	case "Four-Seam Fastball":
		// Fastballs: elevated / top 1/3 of the zone (high heat, slight arm side)
		plateZ = 3.25 + jitterZ
		plateX = (0.18 * armFactor) + jitterX
	case "Sinker":
		// Sinkers: low and arm-side (+X for RHP, -X for LHP)
		plateZ = 1.65 + jitterZ
		plateX = (0.50 * armFactor) + jitterX
	case "Changeup", "Splitter", "Forkball":
		// Changeups / Splitters: low and arm-side fade
		plateZ = 1.50 + jitterZ
		plateX = (0.48 * armFactor) + jitterX
	case "Cutter":
		// Cutters: glove-side high edge (-X for RHP, +X for LHP)
		plateZ = 2.95 + jitterZ
		plateX = (-0.45 * armFactor) + jitterX
	case "Slider":
		// Sliders: low and glove-side (-X for RHP, +X for LHP)
		plateZ = 1.50 + jitterZ
		plateX = (-0.55 * armFactor) + jitterX
	case "Sweeper":
		// Sweepers: low and wide glove-side sweep (-X for RHP, +X for LHP)
		plateZ = 1.60 + jitterZ
		plateX = (-0.75 * armFactor) + jitterX
	case "Curveball", "Slurve":
		// Curveballs / Slurves: low / bottom of the zone (12-6 or looping drop)
		plateZ = 1.35 + jitterZ
		plateX = (-0.12 * armFactor) + jitterX
	default:
		plateZ = 2.20 + jitterZ
		plateX = (0.05 * armFactor) + jitterX
	}

	return plateX, plateZ
}
