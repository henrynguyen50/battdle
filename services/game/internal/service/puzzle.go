package service

import (
	"time"

	"pitchle/services/game/internal/repository"
	"pitchle/shared/models"
)

type PuzzleService struct {
	repo *repository.Repository
}

func NewPuzzleService(repo *repository.Repository) *PuzzleService {
	return &PuzzleService{repo: repo}
}

func (s *PuzzleService) GetTodayPuzzle() (*models.DailyPuzzle, error) {
	// Daily mystery pitcher is selected using UTC time today
	return s.repo.GetOrCreateDailyPuzzle(time.Now().UTC())
}

func (s *PuzzleService) ResetTodayPuzzleForTest(sessionID string) (*models.DailyPuzzle, error) {
	return s.repo.ResetDailyPuzzleForTest(sessionID)
}
func (s *PuzzleService) SetTargetPitcherForTest(playerID int, sessionID string) (*models.DailyPuzzle, error) {
	return s.repo.SetTargetPitcherForTest(playerID, sessionID)
}

type GameRepository interface {
	GetPlayerByID(id int) (*models.Player, error)
	GetPitchProfileByID(id int) (*models.PitchProfile, error)
	GetPitchProfilesByPlayerID(playerID int) ([]models.PitchProfile, error)
	GetGuessesBySessionAndPuzzle(sessionID string, puzzleID int) ([]models.Guess, error)
	HasPlayerBeenGuessed(sessionID string, puzzleID int, playerID int) (bool, error)
	SaveGuess(g *models.Guess) error
	GetPitchGuessBySessionAndPuzzle(sessionID string, puzzleID int) (*models.PitchGuess, error)
	SavePitchGuess(g *models.PitchGuess) error
	ResetDailyPuzzleForTest(sessionID string) (*models.DailyPuzzle, error)
}

type GameService struct {
	repo GameRepository
}

func NewGameService(repo GameRepository) *GameService {
	return &GameService{repo: repo}
}
