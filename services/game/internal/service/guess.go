package service

import (
	"fmt"

	"pitchle/shared/models"
)

type GuessResult struct {
	PlayerID   int                 `json:"player_id"`
	PlayerName string              `json:"player_name"`
	Result     string              `json:"result"` // ball, strike, correct
	Categories CategoryFeedbackMap `json:"categories"`
}

type CategoryFeedback struct {
	Value   interface{} `json:"value"`
	Matched bool        `json:"matched"`
}

type CategoryFeedbackMap struct {
	Team        CategoryFeedback `json:"team"`
	Division    CategoryFeedback `json:"division"`
	YearsPlayed CategoryFeedback `json:"years_played"`
	Position    CategoryFeedback `json:"position"`
	YearBorn    CategoryFeedback `json:"year_born"`
}

type GameState struct {
	Status  string        `json:"status"` // active, won, lost
	Balls   int           `json:"balls"`
	Strikes int           `json:"strikes"`
	Guesses []GuessResult `json:"guesses"`
	Answer  *AnswerPlayer `json:"answer,omitempty"`
}

type AnswerPlayer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *GameService) GetGameState(sessionID string, puzzle *models.DailyPuzzle) (*GameState, error) {
	// Fetch target player metadata
	targetPlayer, err := s.repo.GetPlayerByID(puzzle.TargetPlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target player: %w", err)
	}

	// Fetch all guesses in this session
	guesses, err := s.repo.GetGuessesBySessionAndPuzzle(sessionID, puzzle.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guesses: %w", err)
	}

	var guessResults []GuessResult
	balls := 0
	strikes := 0
	status := "active"

	for _, g := range guesses {
		guessedPlayer, err := s.repo.GetPlayerByID(g.GuessedPlayerID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guessed player: %w", err)
		}

		feedback := CompareCategories(guessedPlayer, targetPlayer)

		guessResults = append(guessResults, GuessResult{
			PlayerID:   g.GuessedPlayerID,
			PlayerName: guessedPlayer.Name,
			Result:     g.Result,
			Categories: feedback,
		})

		if g.Result == "correct" {
			status = "won"
		} else if g.Result == "ball" {
			balls++
		} else if g.Result == "strike" {
			strikes++
		}
	}

	if status != "won" && strikes >= 3 {
		status = "lost"
	}

	state := &GameState{
		Status:  status,
		Balls:   balls,
		Strikes: strikes,
		Guesses: guessResults,
	}

	if status == "won" || status == "lost" {
		state.Answer = &AnswerPlayer{
			ID:   targetPlayer.ID,
			Name: targetPlayer.Name,
		}
	}

	return state, nil
}

func CompareCategories(guessed, target *models.Player) CategoryFeedbackMap {
	guessedYears := guessed.MLBLastYear - guessed.MLBDebutYear + 1
	targetYears := target.MLBLastYear - target.MLBDebutYear + 1

	return CategoryFeedbackMap{
		Team: CategoryFeedback{
			Value:   guessed.TeamName,
			Matched: guessed.TeamID == target.TeamID,
		},
		Division: CategoryFeedback{
			Value:   guessed.DivisionName,
			Matched: guessed.DivisionID == target.DivisionID,
		},
		YearsPlayed: CategoryFeedback{
			Value:   guessedYears,
			Matched: guessedYears == targetYears,
		},
		Position: CategoryFeedback{
			Value:   guessed.Position,
			Matched: guessed.Position == target.Position,
		},
		YearBorn: CategoryFeedback{
			Value:   guessed.BirthYear,
			Matched: guessed.BirthYear == target.BirthYear,
		},
	}
}

func (s *GameService) SubmitGuess(sessionID string, puzzle *models.DailyPuzzle, guessedPlayerID int) (*GameState, error) {
	// 1. Fetch current game state to validate active status
	state, err := s.GetGameState(sessionID, puzzle)
	if err != nil {
		return nil, err
	}

	if state.Status != "active" {
		return nil, fmt.Errorf("game already completed: status is %s", state.Status)
	}

	// 2. Check if player was already guessed in this session
	alreadyGuessed, err := s.repo.HasPlayerBeenGuessed(sessionID, puzzle.ID, guessedPlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if player was guessed: %w", err)
	}
	if alreadyGuessed {
		return nil, fmt.Errorf("player was already guessed in this session")
	}

	// 3. Fetch guessed player and target player metadata
	guessedPlayer, err := s.repo.GetPlayerByID(guessedPlayerID)
	if err != nil {
		return nil, fmt.Errorf("guessed player not found: %w", err)
	}
	targetPlayer, err := s.repo.GetPlayerByID(puzzle.TargetPlayerID)
	if err != nil {
		return nil, fmt.Errorf("target player not found: %w", err)
	}

	// 4. Calculate matching categories and result
	feedback := CompareCategories(guessedPlayer, targetPlayer)

	var result string
	if guessedPlayer.ID == targetPlayer.ID {
		result = "correct"
		state.Status = "won"
	} else {
		// Count matching categories
		matches := 0
		if feedback.Team.Matched {
			matches++
		}
		if feedback.Division.Matched {
			matches++
		}
		if feedback.YearsPlayed.Matched {
			matches++
		}
		if feedback.Position.Matched {
			matches++
		}
		if feedback.YearBorn.Matched {
			matches++
		}

		if matches >= 1 {
			result = "ball"
			state.Balls++
		} else {
			result = "strike"
			state.Strikes++
		}

		if state.Strikes >= 3 {
			state.Status = "lost"
		}
	}

	// 5. Save the guess
	newGuess := &models.Guess{
		SessionID:       sessionID,
		PuzzleID:        puzzle.ID,
		GuessedPlayerID: guessedPlayerID,
		Balls:           state.Balls,
		Strikes:         state.Strikes,
		Result:          result,
	}
	err = s.repo.SaveGuess(newGuess)
	if err != nil {
		return nil, fmt.Errorf("failed to save guess: %w", err)
	}

	// 6. Append new guess to state guesses
	state.Guesses = append(state.Guesses, GuessResult{
		PlayerID:   guessedPlayerID,
		PlayerName: guessedPlayer.Name,
		Result:     result,
		Categories: feedback,
	})

	if state.Status == "won" || state.Status == "lost" {
		state.Answer = &AnswerPlayer{
			ID:   targetPlayer.ID,
			Name: targetPlayer.Name,
		}
	}

	return state, nil
}
