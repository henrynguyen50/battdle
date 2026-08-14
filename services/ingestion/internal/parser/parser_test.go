package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMetadataCSV_Success(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "player_metadata.csv")

	csvContent := `player_id,player_name,birth_date,birth_year,birth_city,birth_country,position,height,weight,mlb_debut_year,mlb_last_year,mlb_team_id
12345,John Doe,1990-01-01,1990,New York,USA,P,"6' 2""",200,2015,2026,147
67890,Jane Smith,1992-05-10,1992,Toronto,Canada,P,"6' 0""",185,2017,2025,148
`
	if err := os.WriteFile(tempFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write temp CSV file: %v", err)
	}

	records, err := ParseMetadataCSV(tempFile)
	if err != nil {
		t.Fatalf("ParseMetadataCSV returned unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Verify player 12345
	r1, ok := records[12345]
	if !ok {
		t.Errorf("expected player 12345 to be in the records")
	} else {
		if r1.PlayerID != 12345 {
			t.Errorf("expected PlayerID 12345, got %d", r1.PlayerID)
		}
		if r1.PlayerName != "John Doe" {
			t.Errorf("expected PlayerName 'John Doe', got %q", r1.PlayerName)
		}
		if r1.BirthDate != "1990-01-01" {
			t.Errorf("expected BirthDate '1990-01-01', got %q", r1.BirthDate)
		}
		if r1.BirthYear != 1990 {
			t.Errorf("expected BirthYear 1990, got %d", r1.BirthYear)
		}
		if r1.BirthCity != "New York" {
			t.Errorf("expected BirthCity 'New York', got %q", r1.BirthCity)
		}
		if r1.BirthCountry != "USA" {
			t.Errorf("expected BirthCountry 'USA', got %q", r1.BirthCountry)
		}
		if r1.Position != "P" {
			t.Errorf("expected Position 'P', got %q", r1.Position)
		}
		if r1.Height != `6' 2"` {
			t.Errorf("expected Height '6' 2\"', got %q", r1.Height)
		}
		if r1.Weight != 200 {
			t.Errorf("expected Weight 200, got %d", r1.Weight)
		}
		if r1.MLBDebutYear != 2015 {
			t.Errorf("expected MLBDebutYear 2015, got %d", r1.MLBDebutYear)
		}
		if r1.MLBLastYear != 2026 {
			t.Errorf("expected MLBLastYear 2026, got %d", r1.MLBLastYear)
		}
		if r1.MLBTeamID != 147 {
			t.Errorf("expected MLBTeamID 147, got %d", r1.MLBTeamID)
		}
	}

	// Verify player 67890
	r2, ok := records[67890]
	if !ok {
		t.Errorf("expected player 67890 to be in the records")
	} else {
		if r2.PlayerID != 67890 {
			t.Errorf("expected PlayerID 67890, got %d", r2.PlayerID)
		}
		if r2.PlayerName != "Jane Smith" {
			t.Errorf("expected PlayerName 'Jane Smith', got %q", r2.PlayerName)
		}
		if r2.BirthDate != "1992-05-10" {
			t.Errorf("expected BirthDate '1992-05-10', got %q", r2.BirthDate)
		}
		if r2.BirthYear != 1992 {
			t.Errorf("expected BirthYear 1992, got %d", r2.BirthYear)
		}
		if r2.BirthCity != "Toronto" {
			t.Errorf("expected BirthCity 'Toronto', got %q", r2.BirthCity)
		}
		if r2.BirthCountry != "Canada" {
			t.Errorf("expected BirthCountry 'Canada', got %q", r2.BirthCountry)
		}
		if r2.Position != "P" {
			t.Errorf("expected Position 'P', got %q", r2.Position)
		}
		if r2.Height != `6' 0"` {
			t.Errorf("expected Height '6' 0\"', got %q", r2.Height)
		}
		if r2.Weight != 185 {
			t.Errorf("expected Weight 185, got %d", r2.Weight)
		}
		if r2.MLBDebutYear != 2017 {
			t.Errorf("expected MLBDebutYear 2017, got %d", r2.MLBDebutYear)
		}
		if r2.MLBLastYear != 2025 {
			t.Errorf("expected MLBLastYear 2025, got %d", r2.MLBLastYear)
		}
		if r2.MLBTeamID != 148 {
			t.Errorf("expected MLBTeamID 148, got %d", r2.MLBTeamID)
		}
	}
}

func TestParseMetadataCSV_ErrorsAndEdges(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ParseMetadataCSV(filepath.Join(tempDir, "does-not-exist.csv"))
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(tempDir, "empty.csv")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write empty file: %v", err)
		}
		_, err := ParseMetadataCSV(path)
		if err == nil {
			t.Errorf("expected error for empty file, got nil")
		}
	})

	t.Run("missing required column", func(t *testing.T) {
		// Missing 'player_name'
		csvContent := `player_id,birth_date,birth_year,birth_city,birth_country,position,height,weight,mlb_debut_year,mlb_last_year,mlb_team_id
12345,1990-01-01,1990,New York,USA,P,"6' 2""",200,2015,2026,147
`
		path := filepath.Join(tempDir, "missing_col.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write missing_col.csv: %v", err)
		}
		_, err := ParseMetadataCSV(path)
		if err == nil {
			t.Errorf("expected error for missing column, got nil")
		}
	})

	t.Run("invalid player id row skipped", func(t *testing.T) {
		csvContent := `player_id,player_name,birth_date,birth_year,birth_city,birth_country,position,height,weight,mlb_debut_year,mlb_last_year,mlb_team_id
not-an-int,John Doe,1990-01-01,1990,New York,USA,P,"6' 2""",200,2015,2026,147
12345,Jane Smith,1992-05-10,1992,Toronto,Canada,P,"6' 0""",185,2017,2025,148
`
		path := filepath.Join(tempDir, "skip_invalid_id.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write skip_invalid_id.csv: %v", err)
		}
		records, err := ParseMetadataCSV(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Errorf("expected 1 record (second row), got %d", len(records))
		}
		if _, ok := records[12345]; !ok {
			t.Errorf("expected player 12345 to be parsed")
		}
	})

	t.Run("malformed csv row", func(t *testing.T) {
		// Row with mismatched quotes or invalid format
		csvContent := `player_id,player_name,birth_date,birth_year,birth_city,birth_country,position,height,weight,mlb_debut_year,mlb_last_year,mlb_team_id
12345,John "Doe,1990-01-01,1990,New York,USA,P,"6' 2""",200,2015,2026,147
`
		path := filepath.Join(tempDir, "malformed_row.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write malformed_row.csv: %v", err)
		}
		_, err := ParseMetadataCSV(path)
		if err == nil {
			t.Errorf("expected error for malformed CSV row, got nil")
		}
	})

	t.Run("invalid optional ints default to 0", func(t *testing.T) {
		csvContent := `player_id,player_name,birth_date,birth_year,birth_city,birth_country,position,height,weight,mlb_debut_year,mlb_last_year,mlb_team_id
12345,John Doe,1990-01-01,not-an-int,New York,USA,P,"6' 2""",abc,xyz,2026,def
`
		path := filepath.Join(tempDir, "invalid_ints.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write invalid_ints.csv: %v", err)
		}
		records, err := ParseMetadataCSV(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r, ok := records[12345]
		if !ok {
			t.Fatalf("expected player 12345 to be in the records")
		}
		if r.BirthYear != 0 {
			t.Errorf("expected BirthYear to default to 0, got %d", r.BirthYear)
		}
		if r.Weight != 0 {
			t.Errorf("expected Weight to default to 0, got %d", r.Weight)
		}
		if r.MLBDebutYear != 0 {
			t.Errorf("expected MLBDebutYear to default to 0, got %d", r.MLBDebutYear)
		}
		if r.MLBLastYear != 2026 {
			t.Errorf("expected MLBLastYear to be parsed as 2026, got %d", r.MLBLastYear)
		}
		if r.MLBTeamID != 0 {
			t.Errorf("expected MLBTeamID to default to 0, got %d", r.MLBTeamID)
		}
	})
}

func TestParseCSV_Success(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "pitch_data.csv")

	csvContent := `player_id,player_name,spin_rate,velocity,release_extension,api_break_x_arm,api_break_z_with_gravity,release_pos_x,release_pos_z,plate_x,plate_z,arm_angle
12345,John Doe,2200.5,95.2,6.5,8.2,-12.4,-1.8,6.1,0.2,2.5,45.0
`
	if err := os.WriteFile(tempFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write temp pitch CSV file: %v", err)
	}

	records, err := ParseCSV(tempFile)
	if err != nil {
		t.Fatalf("ParseCSV returned unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	r := records[0]
	if r.PlayerID != 12345 {
		t.Errorf("expected PlayerID 12345, got %d", r.PlayerID)
	}
	if r.PlayerName != "John Doe" {
		t.Errorf("expected PlayerName 'John Doe', got %q", r.PlayerName)
	}
	if r.SpinRate != 2200.5 {
		t.Errorf("expected SpinRate 2200.5, got %f", r.SpinRate)
	}
	if r.Velocity != 95.2 {
		t.Errorf("expected Velocity 95.2, got %f", r.Velocity)
	}
	if r.ReleaseExtension != 6.5 {
		t.Errorf("expected ReleaseExtension 6.5, got %f", r.ReleaseExtension)
	}
	if r.BreakX != 8.2 {
		t.Errorf("expected BreakX 8.2, got %f", r.BreakX)
	}
	if r.BreakZ != -12.4 {
		t.Errorf("expected BreakZ -12.4, got %f", r.BreakZ)
	}
	if r.ReleasePosX != -1.8 {
		t.Errorf("expected ReleasePosX -1.8, got %f", r.ReleasePosX)
	}
	if r.ReleasePosZ != 6.1 {
		t.Errorf("expected ReleasePosZ 6.1, got %f", r.ReleasePosZ)
	}
	if r.PlateX != 0.2 {
		t.Errorf("expected PlateX 0.2, got %f", r.PlateX)
	}
	if r.PlateZ != 2.5 {
		t.Errorf("expected PlateZ 2.5, got %f", r.PlateZ)
	}
	if r.ArmAngle != 45.0 {
		t.Errorf("expected ArmAngle 45.0, got %f", r.ArmAngle)
	}
}

func TestParseCSV_ErrorsAndEdges(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ParseCSV(filepath.Join(tempDir, "does-not-exist.csv"))
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(tempDir, "empty.csv")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write empty file: %v", err)
		}
		_, err := ParseCSV(path)
		if err == nil {
			t.Errorf("expected error for empty file, got nil")
		}
	})

	t.Run("missing required column", func(t *testing.T) {
		// Missing 'velocity'
		csvContent := `player_id,player_name,spin_rate,release_extension,api_break_x_arm,api_break_z_with_gravity,release_pos_x,release_pos_z,plate_x,plate_z,arm_angle
12345,John Doe,2200.5,6.5,8.2,-12.4,-1.8,6.1,0.2,2.5,45.0
`
		path := filepath.Join(tempDir, "missing_col.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write missing_col.csv: %v", err)
		}
		_, err := ParseCSV(path)
		if err == nil {
			t.Errorf("expected error for missing column, got nil")
		}
	})

	t.Run("invalid player id row skipped", func(t *testing.T) {
		csvContent := `player_id,player_name,spin_rate,velocity,release_extension,api_break_x_arm,api_break_z_with_gravity,release_pos_x,release_pos_z,plate_x,plate_z,arm_angle
not-an-int,John Doe,2200.5,95.2,6.5,8.2,-12.4,-1.8,6.1,0.2,2.5,45.0
12345,Jane Smith,2100.0,92.0,6.0,7.5,-10.0,-1.5,6.0,0.1,2.0,40.0
`
		path := filepath.Join(tempDir, "skip_invalid_id.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write skip_invalid_id.csv: %v", err)
		}
		records, err := ParseCSV(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Errorf("expected 1 record, got %d", len(records))
		}
		if records[0].PlayerID != 12345 {
			t.Errorf("expected player 12345 to be parsed, got %d", records[0].PlayerID)
		}
	})

	t.Run("malformed csv row", func(t *testing.T) {
		csvContent := `player_id,player_name,spin_rate,velocity,release_extension,api_break_x_arm,api_break_z_with_gravity,release_pos_x,release_pos_z,plate_x,plate_z,arm_angle
12345,John "Doe,2200.5,95.2,6.5,8.2,-12.4,-1.8,6.1,0.2,2.5,45.0
`
		path := filepath.Join(tempDir, "malformed_row.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write malformed_row.csv: %v", err)
		}
		_, err := ParseCSV(path)
		if err == nil {
			t.Errorf("expected error for malformed CSV row, got nil")
		}
	})

	t.Run("invalid optional floats default to 0.0", func(t *testing.T) {
		csvContent := `player_id,player_name,spin_rate,velocity,release_extension,api_break_x_arm,api_break_z_with_gravity,release_pos_x,release_pos_z,plate_x,plate_z,arm_angle
12345,John Doe,not-a-float,95.2,abc,-12.4,-1.8,xyz,def,0.2,2.5,45.0
`
		path := filepath.Join(tempDir, "invalid_floats.csv")
		if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
			t.Fatalf("failed to write invalid_floats.csv: %v", err)
		}
		records, err := ParseCSV(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		r := records[0]
		if r.SpinRate != 0.0 {
			t.Errorf("expected SpinRate to default to 0.0, got %f", r.SpinRate)
		}
		if r.Velocity != 95.2 {
			t.Errorf("expected Velocity to be parsed as 95.2, got %f", r.Velocity)
		}
		if r.ReleaseExtension != 0.0 {
			t.Errorf("expected ReleaseExtension to default to 0.0, got %f", r.ReleaseExtension)
		}
		if r.ReleasePosZ != 0.0 {
			t.Errorf("expected ReleasePosZ to default to 0.0, got %f", r.ReleasePosZ)
		}
	})
}
