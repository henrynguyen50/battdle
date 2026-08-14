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
	getPlayerByIDFunc                func(id int) (*models.Player, error)
	getGuessesBySessionAndPuzzleFunc func(sessionID string, puzzleID int) ([]models.Guess, error)
	hasPlayerBeenGuessedFunc         func(sessionID string, puzzleID int, playerID int) (bool, error)
	saveGuessFunc                    func(g *models.Guess) error
}

func (m *mockRepository) GetPlayerByID(id int) (*models.Player, error) {
	if m.getPlayerByIDFunc != nil {
		return m.getPlayerByIDFunc(id)
	}
	return nil, fmt.Errorf("GetPlayerByID not implemented")
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

func TestCompareCategories(t *testing.T) {
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
	}

	tests := []struct {
		name    string
		guessed *models.Player
		want    service.CategoryFeedbackMap
	}{
		{
			name: "all categories match",
			guessed: &models.Player{
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
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Dodgers", Matched: true},
				Division:    service.CategoryFeedback{Value: "NL West", Matched: true},
				YearsPlayed: service.CategoryFeedback{Value: 17, Matched: true},
				Position:    service.CategoryFeedback{Value: "SP", Matched: true},
				YearBorn:    service.CategoryFeedback{Value: 1988, Matched: true},
			},
		},
		{
			name: "only team and division match",
			guessed: &models.Player{
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
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Dodgers", Matched: true},
				Division:    service.CategoryFeedback{Value: "NL West", Matched: true},
				YearsPlayed: service.CategoryFeedback{Value: 15, Matched: false},
				Position:    service.CategoryFeedback{Value: "1B", Matched: false},
				YearBorn:    service.CategoryFeedback{Value: 1989, Matched: false},
			},
		},
		{
			name: "only division matches",
			guessed: &models.Player{
				ID:           3,
				MLBID:        1003,
				Name:         "Logan Webb",
				BirthYear:    1996,
				Position:     "RP",
				MLBDebutYear: 2019,
				MLBLastYear:  2024,
				TeamID:       11,
				TeamName:     "Giants",
				DivisionID:   100,
				DivisionName: "NL West",
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Giants", Matched: false},
				Division:    service.CategoryFeedback{Value: "NL West", Matched: true},
				YearsPlayed: service.CategoryFeedback{Value: 6, Matched: false},
				Position:    service.CategoryFeedback{Value: "RP", Matched: false},
				YearBorn:    service.CategoryFeedback{Value: 1996, Matched: false},
			},
		},
		{
			name: "only years played matches",
			guessed: &models.Player{
				ID:           4,
				MLBID:        1004,
				Name:         "Zack Greinke",
				BirthYear:    1983,
				Position:     "RP",
				MLBDebutYear: 2007,
				MLBLastYear:  2023,
				TeamID:       12,
				TeamName:     "Royals",
				DivisionID:   101,
				DivisionName: "AL Central",
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Royals", Matched: false},
				Division:    service.CategoryFeedback{Value: "AL Central", Matched: false},
				YearsPlayed: service.CategoryFeedback{Value: 17, Matched: true},
				Position:    service.CategoryFeedback{Value: "RP", Matched: false},
				YearBorn:    service.CategoryFeedback{Value: 1983, Matched: false},
			},
		},
		{
			name: "only position matches",
			guessed: &models.Player{
				ID:           5,
				MLBID:        1005,
				Name:         "Gerrit Cole",
				BirthYear:    1990,
				Position:     "SP",
				MLBDebutYear: 2013,
				MLBLastYear:  2024,
				TeamID:       13,
				TeamName:     "Yankees",
				DivisionID:   102,
				DivisionName: "AL East",
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Yankees", Matched: false},
				Division:    service.CategoryFeedback{Value: "AL East", Matched: false},
				YearsPlayed: service.CategoryFeedback{Value: 12, Matched: false},
				Position:    service.CategoryFeedback{Value: "SP", Matched: true},
				YearBorn:    service.CategoryFeedback{Value: 1990, Matched: false},
			},
		},
		{
			name: "only year born matches",
			guessed: &models.Player{
				ID:           6,
				MLBID:        1006,
				Name:         "Craig Kimbrel",
				BirthYear:    1988,
				Position:     "RP",
				MLBDebutYear: 2010,
				MLBLastYear:  2024,
				TeamID:       14,
				TeamName:     "Orioles",
				DivisionID:   102,
				DivisionName: "AL East",
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Orioles", Matched: false},
				Division:    service.CategoryFeedback{Value: "AL East", Matched: false},
				YearsPlayed: service.CategoryFeedback{Value: 15, Matched: false},
				Position:    service.CategoryFeedback{Value: "RP", Matched: false},
				YearBorn:    service.CategoryFeedback{Value: 1988, Matched: true},
			},
		},
		{
			name: "all categories mismatch",
			guessed: &models.Player{
				ID:           7,
				MLBID:        1007,
				Name:         "Shohei Ohtani",
				BirthYear:    1994,
				Position:     "DH",
				MLBDebutYear: 2018,
				MLBLastYear:  2024,
				TeamID:       15,
				TeamName:     "Angels",
				DivisionID:   103,
				DivisionName: "AL West",
			},
			want: service.CategoryFeedbackMap{
				Team:        service.CategoryFeedback{Value: "Angels", Matched: false},
				Division:    service.CategoryFeedback{Value: "AL West", Matched: false},
				YearsPlayed: service.CategoryFeedback{Value: 7, Matched: false},
				Position:    service.CategoryFeedback{Value: "DH", Matched: false},
				YearBorn:    service.CategoryFeedback{Value: 1994, Matched: false},
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
			if got.YearsPlayed != tt.want.YearsPlayed {
				t.Errorf("YearsPlayed feedback = %+v, want %+v", got.YearsPlayed, tt.want.YearsPlayed)
			}
			if got.Position != tt.want.Position {
				t.Errorf("Position feedback = %+v, want %+v", got.Position, tt.want.Position)
			}
			if got.YearBorn != tt.want.YearBorn {
				t.Errorf("YearBorn feedback = %+v, want %+v", got.YearBorn, tt.want.YearBorn)
			}
		})
	}
}

func setupSubmitGuessTest() (*models.DailyPuzzle, *models.Player, map[int]*models.Player) {
	puzzle := &models.DailyPuzzle{
		ID:             100,
		PuzzleDate:     time.Now().UTC(),
		TargetPlayerID: 1,
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
		},
		// Strike player (0 matches)
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
}

func TestSubmitGuess_Ball(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-ball"

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
	state, err := svc.SubmitGuess(sessionID, puzzle, 2) // player 2 has matches -> ball
	if err != nil {
		t.Fatalf("unexpected error submitting guess: %v", err)
	}

	if state.Status != "active" {
		t.Errorf("expected state status to be 'active', got %s", state.Status)
	}
	if state.Balls != 1 {
		t.Errorf("expected balls to be 1, got %d", state.Balls)
	}
	if state.Strikes != 0 {
		t.Errorf("expected strikes to be 0, got %d", state.Strikes)
	}
	if state.Answer != nil {
		t.Errorf("expected answer to be nil for active game, got %+v", state.Answer)
	}
	if len(state.Guesses) != 1 {
		t.Fatalf("expected 1 guess, got %d", len(state.Guesses))
	}
	g := state.Guesses[0]
	if g.Result != "ball" {
		t.Errorf("expected result to be 'ball', got %s", g.Result)
	}

	if savedGuess == nil {
		t.Fatal("expected guess to be saved to repository, but was nil")
	}
	if savedGuess.Result != "ball" {
		t.Errorf("expected saved guess result to be 'ball', got %s", savedGuess.Result)
	}
}

func TestSubmitGuess_Strike(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-strike"

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
	state, err := svc.SubmitGuess(sessionID, puzzle, 3) // player 3 has 0 matches -> strike
	if err != nil {
		t.Fatalf("unexpected error submitting guess: %v", err)
	}

	if state.Status != "active" {
		t.Errorf("expected state status to be 'active', got %s", state.Status)
	}
	if state.Balls != 0 {
		t.Errorf("expected balls to be 0, got %d", state.Balls)
	}
	if state.Strikes != 1 {
		t.Errorf("expected strikes to be 1, got %d", state.Strikes)
	}
	if len(state.Guesses) != 1 {
		t.Fatalf("expected 1 guess, got %d", len(state.Guesses))
	}
	g := state.Guesses[0]
	if g.Result != "strike" {
		t.Errorf("expected result to be 'strike', got %s", g.Result)
	}

	if savedGuess == nil {
		t.Fatal("expected guess to be saved to repository, but was nil")
	}
	if savedGuess.Result != "strike" {
		t.Errorf("expected saved guess result to be 'strike', got %s", savedGuess.Result)
	}
}

func TestSubmitGuess_StrikeToLost(t *testing.T) {
	puzzle, target, players := setupSubmitGuessTest()
	sessionID := "session-lost"

	// Mock existing 2 strike guesses in DB
	existingGuesses := []models.Guess{
		{ID: 10, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 0, Strikes: 1, Result: "strike"},
		{ID: 11, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 0, Strikes: 2, Result: "strike"},
	}

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
			return existingGuesses, nil
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
	state, err := svc.SubmitGuess(sessionID, puzzle, 3) // player 3 has 0 matches -> 3rd strike
	if err != nil {
		t.Fatalf("unexpected error submitting guess: %v", err)
	}

	if state.Status != "lost" {
		t.Errorf("expected state status to be 'lost', got %s", state.Status)
	}
	if state.Balls != 0 {
		t.Errorf("expected balls to be 0, got %d", state.Balls)
	}
	if state.Strikes != 3 {
		t.Errorf("expected strikes to be 3, got %d", state.Strikes)
	}
	if state.Answer == nil || state.Answer.ID != target.ID || state.Answer.Name != target.Name {
		t.Errorf("expected state answer to be %+v, got %+v", target, state.Answer)
	}
	if len(state.Guesses) != 3 {
		t.Fatalf("expected 3 guesses, got %d", len(state.Guesses))
	}
	g := state.Guesses[2]
	if g.Result != "strike" {
		t.Errorf("expected last guess result to be 'strike', got %s", g.Result)
	}

	if savedGuess == nil {
		t.Fatal("expected guess to be saved to repository, but was nil")
	}
	if savedGuess.Result != "strike" {
		t.Errorf("expected saved guess result to be 'strike', got %s", savedGuess.Result)
	}
}

func TestSubmitGuess_ValidationErrors(t *testing.T) {
	puzzle, _, players := setupSubmitGuessTest()
	sessionID := "session-validation"

	t.Run("Game already won", func(t *testing.T) {
		// Mock 1 guess which was correct
		existingGuesses := []models.Guess{
			{ID: 10, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 1, Balls: 0, Strikes: 0, Result: "correct"},
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
		_, err := svc.SubmitGuess(sessionID, puzzle, 2)
		if err == nil {
			t.Error("expected error when game is already won, got nil")
		}
	})

	t.Run("Game already lost", func(t *testing.T) {
		// Mock 3 strike guesses
		existingGuesses := []models.Guess{
			{ID: 10, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 0, Strikes: 1, Result: "strike"},
			{ID: 11, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 0, Strikes: 2, Result: "strike"},
			{ID: 12, SessionID: sessionID, PuzzleID: puzzle.ID, GuessedPlayerID: 3, Balls: 0, Strikes: 3, Result: "strike"},
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
		_, err := svc.SubmitGuess(sessionID, puzzle, 2)
		if err == nil {
			t.Error("expected error when game is already lost, got nil")
		}
	})

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
