package service_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"pitchle/services/game/internal/service"
	"pitchle/shared/models"
)

type mockRepository struct {
	getPlayerByIDFunc                   func(id int) (*models.Player, error)
	getPitchProfileByIDFunc             func(id int) (*models.PitchProfile, error)
	getGuessesBySessionAndPuzzleFunc    func(sessionID string, puzzleID int) ([]models.Guess, error)
	hasPlayerBeenGuessedFunc            func(sessionID string, puzzleID int, playerID int) (bool, error)
	saveGuessFunc                       func(g *models.Guess) error
	getPitchGuessBySessionAndPuzzleFunc func(sessionID string, puzzleID int) (*models.PitchGuess, error)
	savePitchGuessFunc                  func(g *models.PitchGuess) error
	getPitchProfilesByPlayerIDFunc      func(playerID int) ([]models.PitchProfile, error)
	resetDailyPuzzleForTestFunc         func(sessionID string) (*models.DailyPuzzle, error)
	recordGameCompletionFunc            func(sessionID string, puzzleID int, status string, guessCount int, pitchMatched bool, timeTakenSec int) error
	getTodayPuzzleStatsFunc             func(puzzleID int, sessionID string) (*models.DailyStats, error)
	getDailyLeaderboardFunc             func(puzzleID int, limit int) ([]models.LeaderboardEntry, error)
	getStreakLeaderboardFunc            func(limit int) ([]models.StreakLeaderboardEntry, error)
}

func (m *mockRepository) GetPitchProfilesByPlayerID(playerID int) ([]models.PitchProfile, error) {
	if m.getPitchProfilesByPlayerIDFunc != nil {
		return m.getPitchProfilesByPlayerIDFunc(playerID)
	}
	return []models.PitchProfile{
		{
			ID:           1,
			PlayerID:     playerID,
			PitchType:    "Four-Seam Fastball",
			Velocity:     97.5,
			SpinRate:     2400.0,
			UsagePercent: 55.0,
		},
	}, nil
}
func (m *mockRepository) GetPlayerByID(id int) (*models.Player, error) {
	if m.getPlayerByIDFunc != nil {
		return m.getPlayerByIDFunc(id)
	}
	return nil, fmt.Errorf("GetPlayerByID not implemented")
}

func (m *mockRepository) GetPitchProfileByID(id int) (*models.PitchProfile, error) {
	if m.getPitchProfileByIDFunc != nil {
		return m.getPitchProfileByIDFunc(id)
	}
	return &models.PitchProfile{
		ID:        id,
		PitchType: "Four-Seam Fastball",
		Velocity:  97.5,
		SpinRate:  2400.0,
	}, nil
}

func (m *mockRepository) GetGuessesBySessionAndPuzzle(sessionID string, puzzleID int) ([]models.Guess, error) {
	if m.getGuessesBySessionAndPuzzleFunc != nil {
		return m.getGuessesBySessionAndPuzzleFunc(sessionID, puzzleID)
	}
	return nil, nil
}

func (m *mockRepository) HasPlayerBeenGuessed(sessionID string, puzzleID int, playerID int) (bool, error) {
	if m.hasPlayerBeenGuessedFunc != nil {
		return m.hasPlayerBeenGuessedFunc(sessionID, puzzleID, playerID)
	}
	return false, nil
}

func (m *mockRepository) SaveGuess(g *models.Guess) error {
	if m.saveGuessFunc != nil {
		return m.saveGuessFunc(g)
	}
	return nil
}

func (m *mockRepository) GetPitchGuessBySessionAndPuzzle(sessionID string, puzzleID int) (*models.PitchGuess, error) {
	if m.getPitchGuessBySessionAndPuzzleFunc != nil {
		return m.getPitchGuessBySessionAndPuzzleFunc(sessionID, puzzleID)
	}
	return &models.PitchGuess{
		SessionID:        sessionID,
		PuzzleID:         puzzleID,
		GuessedPitchType: "Four-Seam Fastball",
		Matched:          true,
	}, nil
}

func (m *mockRepository) SavePitchGuess(g *models.PitchGuess) error {
	if m.savePitchGuessFunc != nil {
		return m.savePitchGuessFunc(g)
	}
	return nil
}

func (m *mockRepository) ResetDailyPuzzleForTest(sessionID string) (*models.DailyPuzzle, error) {
	if m.resetDailyPuzzleForTestFunc != nil {
		return m.resetDailyPuzzleForTestFunc(sessionID)
	}
	return nil, nil
}

func (m *mockRepository) RecordGameCompletion(sessionID string, puzzleID int, status string, guessCount int, pitchMatched bool, timeTakenSec int) error {
	if m.recordGameCompletionFunc != nil {
		return m.recordGameCompletionFunc(sessionID, puzzleID, status, guessCount, pitchMatched, timeTakenSec)
	}
	return nil
}

func (m *mockRepository) GetTodayPuzzleStats(puzzleID int, sessionID string) (*models.DailyStats, error) {
	if m.getTodayPuzzleStatsFunc != nil {
		return m.getTodayPuzzleStatsFunc(puzzleID, sessionID)
	}
	return &models.DailyStats{
		TotalSolved:   0,
		TotalAttempts: 0,
		WinRate:       0,
		Distribution: models.GuessDistributionMap{
			"1": 0, "2": 0, "3": 0, "4": 0, "5": 0,
			"6": 0, "7": 0, "8": 0, "9+": 0,
		},
		UserStats: &models.UserStats{},
	}, nil
}

func (m *mockRepository) GetDailyLeaderboard(puzzleID int, limit int) ([]models.LeaderboardEntry, error) {
	if m.getDailyLeaderboardFunc != nil {
		return m.getDailyLeaderboardFunc(puzzleID, limit)
	}
	return []models.LeaderboardEntry{}, nil
}

func (m *mockRepository) GetStreakLeaderboard(limit int) ([]models.StreakLeaderboardEntry, error) {
	if m.getStreakLeaderboardFunc != nil {
		return m.getStreakLeaderboardFunc(limit)
	}
	return []models.StreakLeaderboardEntry{}, nil
}

func TestCompareCategories(t *testing.T) {
	target := &models.Player{
		ID:                 1,
		MLBID:              1001,
		Name:               "Clayton Kershaw",
		BirthYear:          1988, // Age: 2026 - 1988 = 38
		BirthCountry:       "USA",
		Height:             "6' 4\"",
		Position:           "SP",
		MLBDebutYear:       2008,
		MLBLastYear:        2024,
		TeamID:             10,
		TeamName:           "Dodgers",
		DivisionID:         100,
		DivisionName:       "NL West",
		League:             "NL",
		PitchHand:          "L",
		KPercent:           25.0,
		BBPercent:          6.0,
		WhiffPercent:       28.0,
		InZonePercent:      48.0,
		GroundballsPercent: 45.0,
		FlyballsPercent:    25.0,
		PopupsPercent:      8.0,
	}

	tests := []struct {
		name    string
		guessed *models.Player
		want    service.CategoryFeedbackMap
	}{
		{
			name: "all categories exact match",
			guessed: &models.Player{
				ID:                 1,
				MLBID:              1001,
				Name:               "Clayton Kershaw",
				BirthYear:          1988,
				BirthCountry:       "USA",
				Height:             "6' 4\"",
				MLBDebutYear:       2008,
				TeamID:             10,
				TeamName:           "Dodgers",
				DivisionID:         100,
				DivisionName:       "NL West",
				League:             "NL",
				PitchHand:          "L",
				KPercent:           25.0,
				BBPercent:          6.0,
				WhiffPercent:       28.0,
				InZonePercent:      48.0,
				GroundballsPercent: 45.0,
				FlyballsPercent:    25.0,
				PopupsPercent:      8.0,
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Dodgers", Matched: true, Close: false, Direction: "equal"},
				Division:    service.CategoryFeedback{Value: "NL West", Matched: true, Close: false, Direction: "equal"},
				Country:     service.CategoryFeedback{Value: "USA", Matched: true, Close: false, Direction: "equal"},
				Height:      service.CategoryFeedback{Value: "6' 4\"", Matched: true, Close: false, Direction: "equal"},
				Age:         service.CategoryFeedback{Value: 38, Matched: true, Close: false, Direction: "equal"},
				Debut:       service.CategoryFeedback{Value: 2008, Matched: true, Close: false, Direction: "equal"},
				Throws:      service.CategoryFeedback{Value: "L", Matched: true, Close: false, Direction: "equal"},
				KPercent:    service.CategoryFeedback{Value: 25.0, Matched: true, Close: false, Direction: "equal"},
				BBPercent:   service.CategoryFeedback{Value: 6.0, Matched: true, Close: false, Direction: "equal"},
				Whiff:       service.CategoryFeedback{Value: 28.0, Matched: true, Close: false, Direction: "equal"},
				InZone:      service.CategoryFeedback{Value: 48.0, Matched: true, Close: false, Direction: "equal"},
				Groundballs: service.CategoryFeedback{Value: 45.0, Matched: true, Close: false, Direction: "equal"},
				Flyballs:    service.CategoryFeedback{Value: 25.0, Matched: true, Close: false, Direction: "equal"},
				Popups:      service.CategoryFeedback{Value: 8.0, Matched: true, Close: false, Direction: "equal"},
			},
		},
		{
			name: "close match on metrics with higher/lower directions",
			guessed: &models.Player{
				ID:                 2,
				MLBID:              1002,
				Name:               "Freddie Freeman",
				BirthYear:          1990, // Age: 36 (target is 38 -> target higher)
				BirthCountry:       "USA",
				Height:             "6' 5\"", // 77 vs 76 -> diff 1 close, target lower
				MLBDebutYear:       2010,    // diff 2 close, target lower
				TeamID:             11,
				TeamName:           "Giants",
				DivisionID:         100,
				DivisionName:       "NL West",
				League:             "NL",
				PitchHand:          "R",   // Mismatch
				KPercent:           22.0,  // diff +3.0 -> close (target higher)
				BBPercent:          8.0,   // diff -2.0 -> close (target lower)
				WhiffPercent:       31.0,  // diff -3.0 -> close (target lower)
				InZonePercent:      51.0,  // diff -3.0 -> close (target lower)
				GroundballsPercent: 42.0,  // diff +3.0 -> close (target higher)
				FlyballsPercent:    28.0,  // diff -3.0 -> close (target lower)
				PopupsPercent:      6.0,   // diff +2.0 -> close (target higher)
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Giants", Matched: false, Close: true, Direction: ""},
				Division:    service.CategoryFeedback{Value: "NL West", Matched: true, Close: false, Direction: "equal"},
				Country:     service.CategoryFeedback{Value: "USA", Matched: true, Close: false, Direction: "equal"},
				Height:      service.CategoryFeedback{Value: "6' 5\"", Matched: false, Close: true, Direction: "lower"},
				Age:         service.CategoryFeedback{Value: 36, Matched: false, Close: true, Direction: "higher"},
				Debut:       service.CategoryFeedback{Value: 2010, Matched: false, Close: true, Direction: "lower"},
				Throws:      service.CategoryFeedback{Value: "R", Matched: false, Close: false, Direction: ""},
				KPercent:    service.CategoryFeedback{Value: 22.0, Matched: false, Close: true, Direction: "higher"},
				BBPercent:   service.CategoryFeedback{Value: 8.0, Matched: false, Close: true, Direction: "lower"},
				Whiff:       service.CategoryFeedback{Value: 31.0, Matched: false, Close: true, Direction: "lower"},
				InZone:      service.CategoryFeedback{Value: 51.0, Matched: false, Close: true, Direction: "lower"},
				Groundballs: service.CategoryFeedback{Value: 42.0, Matched: false, Close: true, Direction: "higher"},
				Flyballs:    service.CategoryFeedback{Value: 28.0, Matched: false, Close: true, Direction: "lower"},
				Popups:      service.CategoryFeedback{Value: 6.0, Matched: false, Close: true, Direction: "higher"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CompareCategories(tt.guessed, target)
			if got.Team != tt.want.Team {
				t.Errorf("Team feedback = %+v, want %+v", got.Team, tt.want.Team)
			}
			if got.Division != tt.want.Division {
				t.Errorf("Division feedback = %+v, want %+v", got.Division, tt.want.Division)
			}
			if got.Country != tt.want.Country {
				t.Errorf("Country feedback = %+v, want %+v", got.Country, tt.want.Country)
			}
			if got.Height != tt.want.Height {
				t.Errorf("Height feedback = %+v, want %+v", got.Height, tt.want.Height)
			}
			if got.Age != tt.want.Age {
				t.Errorf("Age feedback = %+v, want %+v", got.Age, tt.want.Age)
			}
			if got.Debut != tt.want.Debut {
				t.Errorf("Debut feedback = %+v, want %+v", got.Debut, tt.want.Debut)
			}
			if got.Throws != tt.want.Throws {
				t.Errorf("Throws feedback = %+v, want %+v", got.Throws, tt.want.Throws)
			}
			if got.KPercent != tt.want.KPercent {
				t.Errorf("KPercent feedback = %+v, want %+v", got.KPercent, tt.want.KPercent)
			}
			if got.BBPercent != tt.want.BBPercent {
				t.Errorf("BBPercent feedback = %+v, want %+v", got.BBPercent, tt.want.BBPercent)
			}
			if got.Whiff != tt.want.Whiff {
				t.Errorf("Whiff feedback = %+v, want %+v", got.Whiff, tt.want.Whiff)
			}
			if got.InZone != tt.want.InZone {
				t.Errorf("InZone feedback = %+v, want %+v", got.InZone, tt.want.InZone)
			}
			if got.Groundballs != tt.want.Groundballs {
				t.Errorf("Groundballs feedback = %+v, want %+v", got.Groundballs, tt.want.Groundballs)
			}
			if got.Flyballs != tt.want.Flyballs {
				t.Errorf("Flyballs feedback = %+v, want %+v", got.Flyballs, tt.want.Flyballs)
			}
			if got.Popups != tt.want.Popups {
				t.Errorf("Popups feedback = %+v, want %+v", got.Popups, tt.want.Popups)
			}
		})
	}
}

func setupSubmitGuessTest() (*models.DailyPuzzle, *models.Player, map[int]*models.Player) {
	puzzle := &models.DailyPuzzle{
		ID:                   100,
		PuzzleDate:           time.Now().UTC(),
		TargetPlayerID:       1,
		TargetPitchProfileID: 10,
	}

	target := &models.Player{
		ID:           1,
		MLBID:        1001,
		Name:         "Clayton Kershaw",
		BirthYear:    1988,
		Position:     "SP",
		MLBDebutYear: 2008,
		MLBLastYear:  2024,
		TeamID:       10,
		TeamName:     "Dodgers",
		DivisionID:   100,
		DivisionName: "NL West",
		League:       "NL",
		PitchHand:    "L",
		KPercent:     25.0,
		BBPercent:    6.0,
		WhiffPercent: 28.0,
	}

	players := map[int]*models.Player{
		1: target,
		// Ball player (TeamID and DivisionID match)
		2: {
			ID:           2,
			MLBID:        1002,
			Name:         "Freddie Freeman",
			BirthYear:    1989,
			Position:     "1B",
			MLBDebutYear: 2010,
			MLBLastYear:  2024,
			TeamID:       10,
			TeamName:     "Dodgers",
			DivisionID:   100,
			DivisionName: "NL West",
			League:       "NL",
			PitchHand:    "R",
			KPercent:     20.0,
			BBPercent:    8.0,
			WhiffPercent: 24.0,
		},
		// Strike player (0 matches, 0 close)
		3: {
			ID:           3,
			MLBID:        1003,
			Name:         "Shohei Ohtani",
			BirthYear:    1994,
			Position:     "DH",
			MLBDebutYear: 2018,
			MLBLastYear:  2024,
			TeamID:       15,
			TeamName:     "Angels",
			DivisionID:   103,
			DivisionName: "AL West",
			League:       "AL",
			PitchHand:    "R",
			KPercent:     32.0,
			BBPercent:    2.0,
			WhiffPercent: 15.0,
		},
		// Close-only player
		4: {
			ID:           4,
			MLBID:        1004,
			Name:         "Bryce Harper",
			BirthYear:    1992,
			Position:     "1B",
			MLBDebutYear: 2012,
			MLBLastYear:  2024,
			TeamID:       20,
			TeamName:     "Phillies",
			DivisionID:   104,
			DivisionName: "NL East",
			League:       "NL",
			PitchHand:    "R",
			KPercent:     23.0,
			BBPercent:    7.0,
			WhiffPercent: 26.0,
		},
	}

	return puzzle, target, players
}

func TestSubmitGuess_Correct(t *testing.T) {
	puzzle, target, players := setupSubmitGuessTest()
	sessionID := "session-correct"

	var savedGuess *models.Guess
	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			p, ok := players[id]
			if !ok {
				return nil, errors.New("player not found")
			}
			return p, nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return nil, nil
		},
		hasPlayerBeenGuessedFunc: func(sID string, pID int, playerID int) (bool, error) {
			return false, nil
		},
		saveGuessFunc: func(g *models.Guess) error {
			savedGuess = g
			return nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.SubmitGuess(sessionID, puzzle, target.ID)
	if err != nil {
		t.Fatalf("unexpected error submitting guess: %v", err)
	}

	if state.Status != "won" {
		t.Errorf("expected state status to be 'won', got %s", state.Status)
	}
	if state.Answer == nil || state.Answer.ID != target.ID || state.Answer.Name != target.Name {
		t.Errorf("expected state answer to be %+v, got %+v", target, state.Answer)
	}
	if len(state.Guesses) != 1 {
		t.Fatalf("expected 1 guess, got %d", len(state.Guesses))
	}
	g := state.Guesses[0]
	if g.Result != "correct" {
		t.Errorf("expected result to be 'correct', got %s", g.Result)
	}
	if g.PlayerID != target.ID {
		t.Errorf("expected player ID %d, got %d", target.ID, g.PlayerID)
	}

	if savedGuess == nil {
		t.Fatal("expected guess to be saved to repository, but was nil")
	}
	if savedGuess.Result != "correct" {
		t.Errorf("expected saved guess result to be 'correct', got %s", savedGuess.Result)
	}

	// When game is won, all hints are unlocked
	if state.Hints == nil {
		t.Fatal("expected hints to be unlocked when game is won")
	}
	if len(state.Hints.PitchMix) == 0 {
		t.Error("expected pitch mix to be populated")
	}
	if state.Hints.Role == "" {
		t.Error("expected role to be populated")
	}
	if len(state.Hints.PastTeams) == 0 {
		t.Error("expected past teams to be populated")
	}
}

func TestSubmitGuess_Incorrect(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-incorrect"

	var savedGuess *models.Guess
	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			p, ok := players[id]
			if !ok {
				return nil, errors.New("player not found")
			}
			return p, nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return nil, nil
		},
		hasPlayerBeenGuessedFunc: func(sID string, pID int, playerID int) (bool, error) {
			return false, nil
		},
		saveGuessFunc: func(g *models.Guess) error {
			savedGuess = g
			return nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.SubmitGuess(sessionID, puzzle, 2)
	if err != nil {
		t.Fatalf("unexpected error submitting guess: %v", err)
	}

	if state.Status != "active" {
		t.Errorf("expected state status to be 'active', got %s", state.Status)
	}
	if state.Answer != nil {
		t.Errorf("expected answer to be nil for active game, got %+v", state.Answer)
	}
	if len(state.Guesses) != 1 {
		t.Fatalf("expected 1 guess, got %d", len(state.Guesses))
	}

	if savedGuess == nil {
		t.Fatal("expected guess to be saved to repository, but was nil")
	}
}

func TestGetPuzzleAnswer(t *testing.T) {
	puzzle, target, players := setupSubmitGuessTest()
	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getPitchProfileByIDFunc: func(id int) (*models.PitchProfile, error) {
			return &models.PitchProfile{
				ID:        id,
				PitchType: "Slider",
				Velocity:  88.5,
				SpinRate:  2500,
			}, nil
		},
	}

	svc := service.NewGameService(repo)
	ans, err := svc.GetPuzzleAnswer(puzzle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ans.PlayerID != target.ID || ans.PlayerName != target.Name {
		t.Errorf("expected target player %+v, got %+v", target, ans)
	}
	if ans.PitchType != "Slider" {
		t.Errorf("expected pitch type Slider, got %s", ans.PitchType)
	}
}

func TestSubmitGuess_ValidationErrors(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-validation"

	t.Run("Player already guessed in session", func(t *testing.T) {
		repo := &mockRepository{
			getPlayerByIDFunc: func(id int) (*models.Player, error) {
				return players[id], nil
			},
			getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
				return nil, nil
			},
			hasPlayerBeenGuessedFunc: func(sID string, pID int, playerID int) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewGameService(repo)
		_, err := svc.SubmitGuess(sessionID, puzzle, 2)
		if err == nil {
			t.Error("expected error when player has already been guessed, got nil")
		}
	})
}

func TestSubmitGuess_WithoutPitchGuessFails(t *testing.T) {
	puzzle, target, players := setupSubmitGuessTest()
	sessionID := "session-no-pitch"

	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return nil, nil
		},
		getPitchGuessBySessionAndPuzzleFunc: func(sID string, pID int) (*models.PitchGuess, error) {
			return nil, nil // Pitch not guessed yet
		},
	}

	svc := service.NewGameService(repo)
	_, err := svc.SubmitGuess(sessionID, puzzle, target.ID)
	if err == nil {
		t.Fatal("expected error when submitting guess before pitch guess, got nil")
	}
}

func TestSubmitPitchGuess_Match(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-pitch-match"

	var savedPitchGuess *models.PitchGuess
	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getPitchProfileByIDFunc: func(id int) (*models.PitchProfile, error) {
			return &models.PitchProfile{
				ID:        id,
				PitchType: "Slider",
				Velocity:  88.5,
				SpinRate:  2600.0,
			}, nil
		},
		getPitchGuessBySessionAndPuzzleFunc: func(sID string, pID int) (*models.PitchGuess, error) {
			return nil, nil // not guessed yet
		},
		savePitchGuessFunc: func(g *models.PitchGuess) error {
			savedPitchGuess = g
			return nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.SubmitPitchGuess(sessionID, puzzle, "slider")
	if err != nil {
		t.Fatalf("unexpected error submitting pitch guess: %v", err)
	}

	if !state.PitchGuessed {
		t.Error("expected state.PitchGuessed to be true")
	}
	if state.PitchGuess == nil {
		t.Fatal("expected state.PitchGuess to be non-nil")
	}
	if !state.PitchGuess.Matched {
		t.Errorf("expected Matched to be true, got false")
	}
	if state.PitchGuess.ActualType != "Slider" {
		t.Errorf("expected ActualType to be Slider, got %s", state.PitchGuess.ActualType)
	}
	if state.PitchGuess.Velocity != 88.5 {
		t.Errorf("expected Velocity 88.5, got %f", state.PitchGuess.Velocity)
	}
	if savedPitchGuess == nil || !savedPitchGuess.Matched {
		t.Errorf("expected saved pitch guess to be matched")
	}
}

func TestSubmitPitchGuess_Mismatch(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-pitch-mismatch"

	var savedPitchGuess *models.PitchGuess
	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getPitchProfileByIDFunc: func(id int) (*models.PitchProfile, error) {
			return &models.PitchProfile{
				ID:        id,
				PitchType: "Curveball",
				Velocity:  82.0,
				SpinRate:  2800.0,
			}, nil
		},
		getPitchGuessBySessionAndPuzzleFunc: func(sID string, pID int) (*models.PitchGuess, error) {
			return nil, nil // not guessed yet
		},
		savePitchGuessFunc: func(g *models.PitchGuess) error {
			savedPitchGuess = g
			return nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.SubmitPitchGuess(sessionID, puzzle, "Changeup")
	if err != nil {
		t.Fatalf("unexpected error submitting pitch guess: %v", err)
	}

	if !state.PitchGuessed {
		t.Error("expected state.PitchGuessed to be true")
	}
	if state.PitchGuess == nil {
		t.Fatal("expected state.PitchGuess to be non-nil")
	}
	if state.PitchGuess.Matched {
		t.Errorf("expected Matched to be false, got true")
	}
	if state.PitchGuess.ActualType != "Curveball" {
		t.Errorf("expected ActualType to be Curveball, got %s", state.PitchGuess.ActualType)
	}
	if savedPitchGuess == nil || savedPitchGuess.Matched {
		t.Errorf("expected saved pitch guess matched to be false")
	}
}

func TestSubmitPitchGuess_AlreadyGuessed(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-pitch-already"

	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getPitchGuessBySessionAndPuzzleFunc: func(sID string, pID int) (*models.PitchGuess, error) {
			return &models.PitchGuess{
				SessionID:        sessionID,
				PuzzleID:         puzzle.ID,
				GuessedPitchType: "Sinker",
				Matched:          true,
			}, nil
		},
	}

	svc := service.NewGameService(repo)
	_, err := svc.SubmitPitchGuess(sessionID, puzzle, "Sinker")
	if err == nil {
		t.Fatal("expected error when pitch type was already guessed, got nil")
	}
}

func TestSubmitPitchGuess_Empty(t *testing.T) {
	puzzle, _, _ := setupSubmitGuessTest()
	sessionID := "session-pitch-empty"
	repo := &mockRepository{}
	svc := service.NewGameService(repo)
	_, err := svc.SubmitPitchGuess(sessionID, puzzle, "   ")
	if err == nil {
		t.Fatal("expected error for empty pitch type, got nil")
	}
}

func TestMilestoneHints_At3And5Guesses(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-hints"

	t.Run("Hints unlocked progressively", func(t *testing.T) {
		// 0 guesses -> no hints
		repo0 := &mockRepository{
			getPlayerByIDFunc: func(id int) (*models.Player, error) {
				return players[id], nil
			},
			getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
				return nil, nil
			},
		}
		svc0 := service.NewGameService(repo0)
		state0, err := svc0.GetGameState(sessionID, puzzle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state0.Hints != nil {
			t.Errorf("expected no hints for 0 guesses, got %+v", state0.Hints)
		}

		// 3 guesses -> pitch mix unlocked, role & past teams not yet
		guesses3 := []models.Guess{
			{ID: 1, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 2, Balls: 1, Strikes: 0, Result: "ball"},
			{ID: 2, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 1, Strikes: 1, Result: "strike"},
			{ID: 3, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 4, Balls: 2, Strikes: 1, Result: "ball"},
		}
		repo3 := &mockRepository{
			getPlayerByIDFunc: func(id int) (*models.Player, error) {
				return players[id], nil
			},
			getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
				return guesses3, nil
			},
		}
		svc3 := service.NewGameService(repo3)
		state3, err := svc3.GetGameState(sessionID, puzzle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state3.Hints == nil {
			t.Fatal("expected hints to be non-nil for 3 guesses")
		}
		if len(state3.Hints.PitchMix) == 0 {
			t.Errorf("expected PitchMix to be populated for 3 guesses")
		}
		if state3.Hints.Role != "" {
			t.Errorf("expected Role to be empty for 3 guesses, got %s", state3.Hints.Role)
		}
		if len(state3.Hints.PastTeams) != 0 {
			t.Errorf("expected PastTeams to be empty for 3 guesses, got %+v", state3.Hints.PastTeams)
		}

		// 5 guesses -> pitch mix, role, and past teams all unlocked
		guesses5 := []models.Guess{
			{ID: 1, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 2, Balls: 1, Strikes: 0, Result: "ball"},
			{ID: 2, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 1, Strikes: 1, Result: "strike"},
			{ID: 3, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 4, Balls: 2, Strikes: 1, Result: "ball"},
			{ID: 4, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 2, Balls: 3, Strikes: 1, Result: "ball"},
			{ID: 5, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 4, Balls: 4, Strikes: 1, Result: "ball"},
		}
		repo5 := &mockRepository{
			getPlayerByIDFunc: func(id int) (*models.Player, error) {
				return players[id], nil
			},
			getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
				return guesses5, nil
			},
		}
		svc5 := service.NewGameService(repo5)
		state5, err := svc5.GetGameState(sessionID, puzzle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state5.Hints == nil {
			t.Fatal("expected hints to be non-nil for 5 guesses")
		}
		if len(state5.Hints.PitchMix) == 0 {
			t.Errorf("expected PitchMix to be populated for 5 guesses")
		}
		if state5.Hints.Role == "" {
			t.Errorf("expected Role to be populated for 5 guesses")
		}
		if len(state5.Hints.PastTeams) == 0 {
			t.Errorf("expected PastTeams to be populated for 5 guesses")
		}
	})
}

func TestSubmitGuess_RecordsCompletionOnWin(t *testing.T) {
	puzzle, target, players := setupSubmitGuessTest()
	sessionID := "session-win-completion"

	var completedStatus string
	var completedGuessCount int
	var completionRecorded bool

	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return nil, nil
		},
		hasPlayerBeenGuessedFunc: func(sID string, pID int, playerID int) (bool, error) {
			return false, nil
		},
		saveGuessFunc: func(g *models.Guess) error {
			return nil
		},
		recordGameCompletionFunc: func(sID string, pID int, status string, guessCount int, pitchMatched bool, timeTakenSec int) error {
			completionRecorded = true
			completedStatus = status
			completedGuessCount = guessCount
			return nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.SubmitGuess(sessionID, puzzle, target.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Status != "won" {
		t.Errorf("expected won status, got %s", state.Status)
	}
	if !completionRecorded {
		t.Error("expected RecordGameCompletion to be called on win")
	}
	if completedStatus != "won" {
		t.Errorf("expected completedStatus to be 'won', got %s", completedStatus)
	}
	if completedGuessCount != 1 {
		t.Errorf("expected completedGuessCount to be 1, got %d", completedGuessCount)
	}
}

func TestSubmitGuess_RecordsCompletionOnLossAfter9Guesses(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-loss-completion"

	var completedStatus string
	var completedGuessCount int
	var completionRecorded bool

	existingGuesses := make([]models.Guess, 8)
	for i := range 8 {
		existingGuesses[i] = models.Guess{
			ID:              i + 1,
			SessionID:       sessionID,
			PuzzleID:        puzzle.ID,
			GuessedPlayerID: 2,
			Result:          "guess",
		}
	}

	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return existingGuesses, nil
		},
		hasPlayerBeenGuessedFunc: func(sID string, pID int, playerID int) (bool, error) {
			return false, nil
		},
		saveGuessFunc: func(g *models.Guess) error {
			return nil
		},
		recordGameCompletionFunc: func(sID string, pID int, status string, guessCount int, pitchMatched bool, timeTakenSec int) error {
			completionRecorded = true
			completedStatus = status
			completedGuessCount = guessCount
			return nil
		},
	}

	svc := service.NewGameService(repo)
	// Guess an incorrect player (ID 3) as 9th guess
	state, err := svc.SubmitGuess(sessionID, puzzle, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Status != "lost" {
		t.Errorf("expected lost status after 9 guesses, got %s", state.Status)
	}
	if !completionRecorded {
		t.Error("expected RecordGameCompletion to be called on loss")
	}
	if completedStatus != "lost" {
		t.Errorf("expected completedStatus to be 'lost', got %s", completedStatus)
	}
	if completedGuessCount != 9 {
		t.Errorf("expected completedGuessCount to be 9, got %d", completedGuessCount)
	}
}

func TestGetGameState_StatusLostAfter9Guesses(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-state-lost"

	existingGuesses := make([]models.Guess, 9)
	for i := range 9 {
		existingGuesses[i] = models.Guess{
			ID:              i + 1,
			SessionID:       sessionID,
			PuzzleID:        puzzle.ID,
			GuessedPlayerID: 2,
			Result:          "guess",
		}
	}

	repo := &mockRepository{
		getPlayerByIDFunc: func(id int) (*models.Player, error) {
			return players[id], nil
		},
		getGuessesBySessionAndPuzzleFunc: func(sID string, pID int) ([]models.Guess, error) {
			return existingGuesses, nil
		},
	}

	svc := service.NewGameService(repo)
	state, err := svc.GetGameState(sessionID, puzzle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Status != "lost" {
		t.Errorf("expected state status to be 'lost' when 9 guesses exist, got %s", state.Status)
	}
}
