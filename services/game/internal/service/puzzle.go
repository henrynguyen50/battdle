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

type GameRepository interface {
	GetPlayerByID(id int) (*models.Player, error)
	GetGuessesBySessionAndPuzzle(sessionID string, puzzleID int) ([]models.Guess, error)
	HasPlayerBeenGuessed(sessionID string, puzzleID int, playerID int) (bool, error)
	SaveGuess(g *models.Guess) error
}

type GameService struct {
	repo GameRepository
}

func NewGameService(repo GameRepository) *GameService {
	return &GameService{repo: repo}
}
