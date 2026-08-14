package database_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrationFilesExist verifies all required schema migration files exist and are not empty.
func TestMigrationFilesExist(t *testing.T) {
	// Locate repository root from current test directory
	repoRoot := "."
	for range 3 {
		if _, err := os.Stat(filepath.Join(repoRoot, "database", "migrations")); err == nil {
			break
		}
		repoRoot = filepath.Join("..", repoRoot)
	}

	migrationsDir := filepath.Join(repoRoot, "database", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		t.Fatalf("migrations directory not found: %s", migrationsDir)
	}

	requiredFiles := []string{
		"000001_init_schema.up.sql",
		"000001_init_schema.down.sql",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read migration file %s: %v", file, err)
			continue
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			t.Errorf("migration file %s is empty", file)
		}
	}
}

// TestUpMigrationContainsRequiredTables checks that all core application tables are created.
func TestUpMigrationContainsRequiredTables(t *testing.T) {
	repoRoot := "."
	for range 3 {
		if _, err := os.Stat(filepath.Join(repoRoot, "database", "migrations")); err == nil {
			break
		}
		repoRoot = filepath.Join("..", repoRoot)
	}

	path := filepath.Join(repoRoot, "database", "migrations", "000001_init_schema.up.sql")
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read 000001_init_schema.up.sql: %v", err)
	}

	sql := string(contentBytes)
	requiredTables := []string{
		"divisions",
		"teams",
		"players",
		"pitch_profiles",
		"daily_puzzles",
		"guesses",
		"pitch_guesses",
		"animations",
		"game_completions",
		"user_streaks",
	}

	for _, table := range requiredTables {
		if !strings.Contains(sql, "CREATE TABLE "+table) && !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Errorf("000001_init_schema.up.sql missing CREATE TABLE statement for '%s'", table)
		}
	}
}

// TestDownMigrationDropsRequiredTables checks that the down migration drops all tables cleanly.
func TestDownMigrationDropsRequiredTables(t *testing.T) {
	repoRoot := "."
	for range 3 {
		if _, err := os.Stat(filepath.Join(repoRoot, "database", "migrations")); err == nil {
			break
		}
		repoRoot = filepath.Join("..", repoRoot)
	}

	path := filepath.Join(repoRoot, "database", "migrations", "000001_init_schema.down.sql")
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read 000001_init_schema.down.sql: %v", err)
	}

	sql := string(contentBytes)
	requiredDrops := []string{
		"game_completions",
		"user_streaks",
		"animations",
		"pitch_guesses",
		"guesses",
		"daily_puzzles",
		"pitch_profiles",
		"players",
		"teams",
		"divisions",
	}

	for _, table := range requiredDrops {
		if !strings.Contains(sql, "DROP TABLE IF EXISTS "+table) && !strings.Contains(sql, "DROP TABLE "+table) {
			t.Errorf("000001_init_schema.down.sql missing DROP TABLE statement for '%s'", table)
		}
	}
}
