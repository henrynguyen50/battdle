package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"pitchle/shared/models"
)

type GuessResult struct {
	PlayerID   int                 `json:"player_id"`
	PlayerName string              `json:"player_name"`
	Result     string              `json:"result"` // ball, strike, correct
	Categories CategoryFeedbackMap `json:"categories"`
}

type CategoryFeedback struct {
	Value     interface{} `json:"value"`
	Matched   bool        `json:"matched"`
	Close     bool        `json:"close"`
	Direction string      `json:"direction"` // "higher", "lower", or "equal"
}

type CategoryFeedbackMap struct {
	Team        CategoryFeedback `json:"team"`
	Division    CategoryFeedback `json:"division"`
	Country     CategoryFeedback `json:"country"`
	Height      CategoryFeedback `json:"height"`
	Age         CategoryFeedback `json:"age"`
	Debut       CategoryFeedback `json:"debut"`
	Throws      CategoryFeedback `json:"throws"`
	KPercent    CategoryFeedback `json:"k_percent"`
	BBPercent   CategoryFeedback `json:"bb_percent"`
	Whiff       CategoryFeedback `json:"whiff"`
	InZone      CategoryFeedback `json:"in_zone"`
	Groundballs CategoryFeedback `json:"groundballs"`
	Flyballs    CategoryFeedback `json:"flyballs"`
	Popups      CategoryFeedback `json:"popups"`
}
type PitchFeedback struct {
	GuessedType string  `json:"guessed_type"`
	ActualType  string  `json:"actual_type"`
	Matched     bool    `json:"matched"`
	Velocity    float64 `json:"velocity"`
	SpinRate    float64 `json:"spin_rate"`
}

type PitcherHints struct {
	PitchMix  []string `json:"pitch_mix,omitempty"`  // Unlocked when len(guesses) >= 3 or game completed
	Role      string   `json:"role,omitempty"`       // Unlocked when len(guesses) >= 5 or game completed ("Starting Pitcher (SP)" or "Relief Pitcher (RP/CP)")
	PastTeams []string `json:"past_teams,omitempty"` // Unlocked when len(guesses) >= 5 or game completed
}

type GameState struct {
	Status       string         `json:"status"` // active, won, lost
	Balls        int            `json:"balls"`
	Strikes      int            `json:"strikes"`
	PitchGuessed bool           `json:"pitch_guessed"`
	PitchGuess   *PitchFeedback `json:"pitch_guess,omitempty"`
	Guesses      []GuessResult  `json:"guesses"`
	Hints        *PitcherHints  `json:"hints,omitempty"`
	Answer       *AnswerPlayer  `json:"answer,omitempty"`
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

	targetPitchProfile, _ := s.repo.GetPitchProfileByID(puzzle.TargetPitchProfileID)

	// Fetch all guesses in this session
	guesses, err := s.repo.GetGuessesBySessionAndPuzzle(sessionID, puzzle.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guesses: %w", err)
	}

	var guessResults []GuessResult
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
		}
	}

	if status != "won" && len(guessResults) >= 9 {
		status = "lost"
	}

	state := &GameState{
		Status:  status,
		Balls:   0,
		Strikes: 0,
		Guesses: guessResults,
	}
	// Check pitch guess
	pitchGuess, err := s.repo.GetPitchGuessBySessionAndPuzzle(sessionID, puzzle.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pitch guess: %w", err)
	}
	if pitchGuess != nil && targetPitchProfile != nil {
		state.PitchGuessed = true
		state.PitchGuess = &PitchFeedback{
			GuessedType: pitchGuess.GuessedPitchType,
			ActualType:  targetPitchProfile.PitchType,
			Matched:     pitchGuess.Matched,
			Velocity:    targetPitchProfile.Velocity,
			SpinRate:    targetPitchProfile.SpinRate,
		}
	}

	isGameOver := status == "won" || status == "lost"
	targetProfiles, _ := s.repo.GetPitchProfilesByPlayerID(targetPlayer.ID)
	state.Hints = generateHints(targetPlayer, targetPitchProfile, targetProfiles, len(guessResults), isGameOver)

	if isGameOver {
		state.Answer = &AnswerPlayer{
			ID:   targetPlayer.ID,
			Name: targetPlayer.Name,
		}
	}

	return state, nil
}

func getLeague(divOrLeague string) string {
	if strings.HasPrefix(divOrLeague, "AL") {
		return "AL"
	}
	if strings.HasPrefix(divOrLeague, "NL") {
		return "NL"
	}
	return divOrLeague
}

func resolveLeague(p *models.Player) string {
	if p.League != "" {
		return getLeague(p.League)
	}
	return getLeague(p.DivisionName)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
func parseHeightInches(h string) int {
	h = strings.Trim(strings.TrimSpace(h), "\"")
	parts := strings.Split(h, "'")
	if len(parts) >= 2 {
		var feet, inches int
		_, _ = fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &feet)
		inchStr := strings.Trim(strings.TrimSpace(parts[1]), "\" ")
		_, _ = fmt.Sscanf(inchStr, "%d", &inches)
		if feet > 0 {
			return feet*12 + inches
		}
	}
	return 0
}
func CompareCategories(guessed, target *models.Player) CategoryFeedbackMap {
	guessedAge := 2026 - guessed.BirthYear
	targetAge := 2026 - target.BirthYear

	guessedLeague := resolveLeague(guessed)
	targetLeague := resolveLeague(target)

	// Team
	teamMatched := (guessed.TeamID != 0 && target.TeamID != 0 && guessed.TeamID == target.TeamID) ||
		(guessed.TeamName != "" && target.TeamName != "" && guessed.TeamName == target.TeamName)
	teamClose := !teamMatched && (
		(guessed.DivisionID != 0 && target.DivisionID != 0 && guessed.DivisionID == target.DivisionID) ||
			(guessed.DivisionName != "" && target.DivisionName != "" && guessed.DivisionName == target.DivisionName) ||
			(guessedLeague != "" && targetLeague != "" && guessedLeague == targetLeague))

	teamDir := ""
	if teamMatched {
		teamDir = "equal"
	}

	// Division
	divMatched := (guessed.DivisionID != 0 && target.DivisionID != 0 && guessed.DivisionID == target.DivisionID) ||
		(guessed.DivisionName != "" && target.DivisionName != "" && guessed.DivisionName == target.DivisionName)
	divClose := !divMatched && (guessedLeague != "" && targetLeague != "" && guessedLeague == targetLeague)

	divDir := ""
	if divMatched {
		divDir = "equal"
	}

	// Age (Direction: target relative to guessed)
	ageMatched := guessedAge == targetAge
	ageClose := !ageMatched && abs(guessedAge-targetAge) <= 2
	ageDir := "equal"
	if targetAge > guessedAge {
		ageDir = "higher"
	} else if targetAge < guessedAge {
		ageDir = "lower"
	}

	// Throws
	gHand := strings.ToUpper(strings.TrimSpace(guessed.PitchHand))
	if gHand == "" {
		gHand = "R"
	}
	tHand := strings.ToUpper(strings.TrimSpace(target.PitchHand))
	if tHand == "" {
		tHand = "R"
	}
	throwsMatched := gHand == tHand
	throwsDir := ""
	if throwsMatched {
		throwsDir = "equal"
	}

	// K% (Direction: target relative to guessed)
	kDiff := target.KPercent - guessed.KPercent
	kMatched := math.Abs(kDiff) <= 1.5
	kClose := !kMatched && math.Abs(kDiff) <= 4.0
	kDir := "equal"
	if kDiff > 0.05 {
		kDir = "higher"
	} else if kDiff < -0.05 {
		kDir = "lower"
	}

	// BB% (Direction: target relative to guessed)
	bbDiff := target.BBPercent - guessed.BBPercent
	bbMatched := math.Abs(bbDiff) <= 1.0
	bbClose := !bbMatched && math.Abs(bbDiff) <= 2.5
	bbDir := "equal"
	if bbDiff > 0.05 {
		bbDir = "higher"
	} else if bbDiff < -0.05 {
		bbDir = "lower"
	}

	// Whiff% (Direction: target relative to guessed)
	whiffDiff := target.WhiffPercent - guessed.WhiffPercent
	whiffMatched := math.Abs(whiffDiff) <= 1.5
	whiffClose := !whiffMatched && math.Abs(whiffDiff) <= 4.0
	whiffDir := "equal"
	if whiffDiff > 0.05 {
		whiffDir = "higher"
	} else if whiffDiff < -0.05 {
		whiffDir = "lower"
	}

	// Country
	gCountry := strings.TrimSpace(guessed.BirthCountry)
	tCountry := strings.TrimSpace(target.BirthCountry)
	countryMatched := gCountry != "" && tCountry != "" && strings.EqualFold(gCountry, tCountry)
	countryDir := ""
	if countryMatched {
		countryDir = "equal"
	}

	// Height
	gH := parseHeightInches(guessed.Height)
	tH := parseHeightInches(target.Height)
	hMatched := gH > 0 && tH > 0 && gH == tH
	hClose := !hMatched && gH > 0 && tH > 0 && abs(tH-gH) <= 2
	hDir := ""
	if gH > 0 && tH > 0 {
		if tH > gH {
			hDir = "higher"
		} else if tH < gH {
			hDir = "lower"
		} else {
			hDir = "equal"
		}
	}

	// Debut Year
	gDebut := guessed.MLBDebutYear
	tDebut := target.MLBDebutYear
	debutMatched := gDebut > 0 && tDebut > 0 && gDebut == tDebut
	debutClose := !debutMatched && gDebut > 0 && tDebut > 0 && abs(tDebut-gDebut) <= 2
	debutDir := ""
	if gDebut > 0 && tDebut > 0 {
		if tDebut > gDebut {
			debutDir = "higher"
		} else if tDebut < gDebut {
			debutDir = "lower"
		} else {
			debutDir = "equal"
		}
	}

	// In-Zone%
	izDiff := target.InZonePercent - guessed.InZonePercent
	izMatched := math.Abs(izDiff) <= 1.5
	izClose := !izMatched && math.Abs(izDiff) <= 4.0
	izDir := "equal"
	if izDiff > 0.05 {
		izDir = "higher"
	} else if izDiff < -0.05 {
		izDir = "lower"
	}

	// Groundballs%
	gbDiff := target.GroundballsPercent - guessed.GroundballsPercent
	gbMatched := math.Abs(gbDiff) <= 2.0
	gbClose := !gbMatched && math.Abs(gbDiff) <= 5.0
	gbDir := "equal"
	if gbDiff > 0.05 {
		gbDir = "higher"
	} else if gbDiff < -0.05 {
		gbDir = "lower"
	}

	// Flyballs%
	fbDiff := target.FlyballsPercent - guessed.FlyballsPercent
	fbMatched := math.Abs(fbDiff) <= 2.0
	fbClose := !fbMatched && math.Abs(fbDiff) <= 5.0
	fbDir := "equal"
	if fbDiff > 0.05 {
		fbDir = "higher"
	} else if fbDiff < -0.05 {
		fbDir = "lower"
	}

	// Popups%
	popDiff := target.PopupsPercent - guessed.PopupsPercent
	popMatched := math.Abs(popDiff) <= 1.5
	popClose := !popMatched && math.Abs(popDiff) <= 3.5
	popDir := "equal"
	if popDiff > 0.05 {
		popDir = "higher"
	} else if popDiff < -0.05 {
		popDir = "lower"
	}

	displayHeight := strings.TrimSpace(guessed.Height)
	if displayHeight == "" {
		displayHeight = "--"
	}

	displayCountry := guessed.BirthCountry
	if displayCountry == "" {
		displayCountry = "--"
	}

	displayDebut := interface{}(guessed.MLBDebutYear)
	if guessed.MLBDebutYear <= 0 {
		displayDebut = "--"
	}

	return CategoryFeedbackMap{
		Team: CategoryFeedback{
			Value:     guessed.TeamName,
			Matched:   teamMatched,
			Close:     teamClose,
			Direction: teamDir,
		},
		Division: CategoryFeedback{
			Value:     guessed.DivisionName,
			Matched:   divMatched,
			Close:     divClose,
			Direction: divDir,
		},
		Country: CategoryFeedback{
			Value:     displayCountry,
			Matched:   countryMatched,
			Close:     false,
			Direction: countryDir,
		},
		Height: CategoryFeedback{
			Value:     displayHeight,
			Matched:   hMatched,
			Close:     hClose,
			Direction: hDir,
		},
		Age: CategoryFeedback{
			Value:     guessedAge,
			Matched:   ageMatched,
			Close:     ageClose,
			Direction: ageDir,
		},
		Debut: CategoryFeedback{
			Value:     displayDebut,
			Matched:   debutMatched,
			Close:     debutClose,
			Direction: debutDir,
		},
		Throws: CategoryFeedback{
			Value:     gHand,
			Matched:   throwsMatched,
			Close:     false,
			Direction: throwsDir,
		},
		KPercent: CategoryFeedback{
			Value:     guessed.KPercent,
			Matched:   kMatched,
			Close:     kClose,
			Direction: kDir,
		},
		BBPercent: CategoryFeedback{
			Value:     guessed.BBPercent,
			Matched:   bbMatched,
			Close:     bbClose,
			Direction: bbDir,
		},
		Whiff: CategoryFeedback{
			Value:     guessed.WhiffPercent,
			Matched:   whiffMatched,
			Close:     whiffClose,
			Direction: whiffDir,
		},
		InZone: CategoryFeedback{
			Value:     guessed.InZonePercent,
			Matched:   izMatched,
			Close:     izClose,
			Direction: izDir,
		},
		Groundballs: CategoryFeedback{
			Value:     guessed.GroundballsPercent,
			Matched:   gbMatched,
			Close:     gbClose,
			Direction: gbDir,
		},
		Flyballs: CategoryFeedback{
			Value:     guessed.FlyballsPercent,
			Matched:   fbMatched,
			Close:     fbClose,
			Direction: fbDir,
		},
		Popups: CategoryFeedback{
			Value:     guessed.PopupsPercent,
			Matched:   popMatched,
			Close:     popClose,
			Direction: popDir,
		},
	}
}

var mlbTeamList = []string{
	"Baltimore Orioles", "Boston Red Sox", "New York Yankees", "Tampa Bay Rays", "Toronto Blue Jays",
	"Chicago White Sox", "Cleveland Guardians", "Detroit Tigers", "Kansas City Royals", "Minnesota Twins",
	"Houston Astros", "Los Angeles Angels", "Oakland Athletics", "Seattle Mariners", "Texas Rangers",
	"Atlanta Braves", "Miami Marlins", "New York Mets", "Philadelphia Phillies", "Washington Nationals",
	"Chicago Cubs", "Cincinnati Reds", "Milwaukee Brewers", "Pittsburgh Pirates", "St. Louis Cardinals",
	"Arizona Diamondbacks", "Colorado Rockies", "Los Angeles Dodgers", "San Diego Padres", "San Francisco Giants",
}

func getPitchMix(primaryPitch string, playerID int) []string {
	primary := strings.TrimSpace(primaryPitch)
	if primary == "" {
		primary = "Four-Seam Fastball"
	}

	switch primary {
	case "Four-Seam Fastball":
		if playerID%2 == 0 {
			return []string{"Four-Seam Fastball", "Slider", "Changeup", "Curveball"}
		}
		return []string{"Four-Seam Fastball", "Cutter", "Slider", "Changeup"}
	case "Sinker":
		return []string{"Sinker", "Slider", "Changeup", "Sweeper"}
	case "Cutter":
		return []string{"Cutter", "Four-Seam Fastball", "Slider", "Curveball"}
	case "Slider":
		return []string{"Slider", "Four-Seam Fastball", "Sweeper", "Changeup"}
	case "Sweeper":
		return []string{"Sweeper", "Sinker", "Slider", "Changeup"}
	case "Curveball":
		return []string{"Curveball", "Four-Seam Fastball", "Changeup", "Slider"}
	case "Changeup":
		return []string{"Changeup", "Four-Seam Fastball", "Sinker", "Slider"}
	case "Splitter":
		return []string{"Splitter", "Four-Seam Fastball", "Slider", "Curveball"}
	default:
		return []string{primary, "Slider", "Changeup", "Curveball"}
	}
}

func getPlayerRole(p *models.Player) string {
	if strings.Contains(p.Position, "RP") || strings.Contains(p.Position, "CP") || p.Position == "Reliever" {
		return "Relief Pitcher (RP/CP)"
	}
	return "Starting Pitcher (SP)"
}

func getPastTeams(p *models.Player) []string {
	currentTeam := strings.TrimSpace(p.TeamName)
	yearsPlayed := p.MLBLastYear - p.MLBDebutYear + 1
	if yearsPlayed <= 0 {
		yearsPlayed = 1
	}

	if yearsPlayed <= 2 || currentTeam == "" {
		if currentTeam != "" {
			return []string{currentTeam}
		}
		return []string{"Free Agent"}
	}

	// Deterministically pick 1-2 prior teams
	numPastTeams := 1
	if yearsPlayed > 5 {
		numPastTeams = 2
	}

	var pastTeams []string
	idx := (p.ID * 7) % len(mlbTeamList)
	for len(pastTeams) < numPastTeams {
		team := mlbTeamList[idx%len(mlbTeamList)]
		if team != currentTeam {
			exists := false
			for _, pt := range pastTeams {
				if pt == team {
					exists = true
					break
				}
			}
			if !exists {
				pastTeams = append(pastTeams, team)
			}
		}
		idx++
	}

	pastTeams = append(pastTeams, currentTeam)
	return pastTeams
}

func formatPitchArsenal(profiles []models.PitchProfile, fallbackPrimary string, playerID int) []string {
	if len(profiles) > 0 {
		var mix []string
		for _, p := range profiles {
			if p.UsagePercent > 0 {
				mix = append(mix, fmt.Sprintf("%s (%.1f mph, %.1f%%)", p.PitchType, p.Velocity, p.UsagePercent))
			} else {
				mix = append(mix, fmt.Sprintf("%s (%.1f mph)", p.PitchType, p.Velocity))
			}
		}
		return mix
	}
	return getPitchMix(fallbackPrimary, playerID)
}

func generateHints(targetPlayer *models.Player, targetProfile *models.PitchProfile, targetProfiles []models.PitchProfile, numGuesses int, isGameOver bool) *PitcherHints {
	showPitchMix := numGuesses >= 3 || isGameOver
	showRoleAndTeams := numGuesses >= 5 || isGameOver

	if !showPitchMix && !showRoleAndTeams {
		return nil
	}

	hints := &PitcherHints{}

	if showPitchMix {
		primary := "Four-Seam Fastball"
		if targetProfile != nil && targetProfile.PitchType != "" {
			primary = targetProfile.PitchType
		}
		hints.PitchMix = formatPitchArsenal(targetProfiles, primary, targetPlayer.ID)
	}

	if showRoleAndTeams {
		hints.Role = getPlayerRole(targetPlayer)
		hints.PastTeams = getPastTeams(targetPlayer)
	}

	return hints
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

	if !state.PitchGuessed {
		return nil, fmt.Errorf("must guess pitch type before guessing pitcher")
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
		result = "guess"
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
	if state.Status != "won" && len(state.Guesses) >= 9 {
		state.Status = "lost"
	}

	isGameOver := state.Status == "won" || state.Status == "lost"
	if isGameOver {
		pitchMatched := false
		if state.PitchGuess != nil && state.PitchGuess.Matched {
			pitchMatched = true
		} else {
			pg, _ := s.repo.GetPitchGuessBySessionAndPuzzle(sessionID, puzzle.ID)
			if pg != nil && pg.Matched {
				pitchMatched = true
			}
		}

		timeTakenSec := 0
		allGuesses, _ := s.repo.GetGuessesBySessionAndPuzzle(sessionID, puzzle.ID)
		if len(allGuesses) > 0 {
			timeTakenSec = int(time.Since(allGuesses[0].CreatedAt).Seconds())
		}
		if timeTakenSec <= 0 {
			timeTakenSec = 5
		}

		_ = s.repo.RecordGameCompletion(sessionID, puzzle.ID, state.Status, len(state.Guesses), pitchMatched, timeTakenSec)
	}

	targetPitchProfile, _ := s.repo.GetPitchProfileByID(puzzle.TargetPitchProfileID)
	targetProfiles, _ := s.repo.GetPitchProfilesByPlayerID(targetPlayer.ID)
	state.Hints = generateHints(targetPlayer, targetPitchProfile, targetProfiles, len(state.Guesses), isGameOver)
	if isGameOver {
		state.Answer = &AnswerPlayer{
			ID:   targetPlayer.ID,
			Name: targetPlayer.Name,
		}
	}

	return state, nil
}

type TargetPitcherAnswer struct {
	PlayerID           int      `json:"player_id"`
	PlayerName         string   `json:"player_name"`
	TeamName           string   `json:"team_name"`
	DivisionName       string   `json:"division_name"`
	Position           string   `json:"position"`
	Age                int      `json:"age"`
	BirthYear          int      `json:"birth_year"`
	DebutYear          int      `json:"debut_year"`
	LastYear           int      `json:"last_year"`
	PitchHand          string   `json:"pitch_hand"`
	ArmAngle           float64  `json:"arm_angle"`
	KPercent           float64  `json:"k_percent"`
	BBPercent          float64  `json:"bb_percent"`
	WhiffPercent       float64  `json:"whiff_percent"`
	InZonePercent      float64  `json:"in_zone_percent"`
	GroundballsPercent float64  `json:"groundballs_percent"`
	PitchType          string   `json:"pitch_type"`
	Velocity           float64  `json:"velocity"`
	SpinRate           float64  `json:"spin_rate"`
	PitchMix           []string `json:"pitch_mix"`
	Role               string   `json:"role"`
	PastTeams          []string `json:"past_teams"`
}

func (s *GameService) GetPuzzleAnswer(puzzle *models.DailyPuzzle) (*TargetPitcherAnswer, error) {
	targetPlayer, err := s.repo.GetPlayerByID(puzzle.TargetPlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target player: %w", err)
	}

	targetPitchProfile, _ := s.repo.GetPitchProfileByID(puzzle.TargetPitchProfileID)
	pitchType := "Four-Seam Fastball"
	var velo, spin float64
	if targetPitchProfile != nil {
		pitchType = targetPitchProfile.PitchType
		velo = targetPitchProfile.Velocity
		spin = targetPitchProfile.SpinRate
	}

	profiles, _ := s.repo.GetPitchProfilesByPlayerID(targetPlayer.ID)
	pitchMix := formatPitchArsenal(profiles, pitchType, targetPlayer.ID)

	hand := targetPlayer.PitchHand
	if hand == "" {
		hand = "R"
	}

	return &TargetPitcherAnswer{
		PlayerID:           targetPlayer.ID,
		PlayerName:         targetPlayer.Name,
		TeamName:           targetPlayer.TeamName,
		DivisionName:       targetPlayer.DivisionName,
		Position:           targetPlayer.Position,
		Age:                2026 - targetPlayer.BirthYear,
		BirthYear:          targetPlayer.BirthYear,
		DebutYear:          targetPlayer.MLBDebutYear,
		LastYear:           targetPlayer.MLBLastYear,
		PitchHand:          hand,
		ArmAngle:           targetPlayer.ArmAngle,
		KPercent:           targetPlayer.KPercent,
		BBPercent:          targetPlayer.BBPercent,
		WhiffPercent:       targetPlayer.WhiffPercent,
		InZonePercent:      targetPlayer.InZonePercent,
		GroundballsPercent: targetPlayer.GroundballsPercent,
		PitchType:          pitchType,
		Velocity:           velo,
		SpinRate:           spin,
		PitchMix:           pitchMix,
		Role:               getPlayerRole(targetPlayer),
		PastTeams:          getPastTeams(targetPlayer),
	}, nil
}

func (s *GameService) SubmitPitchGuess(sessionID string, puzzle *models.DailyPuzzle, pitchType string) (*GameState, error) {
	pitchType = strings.TrimSpace(pitchType)
	if pitchType == "" {
		return nil, fmt.Errorf("pitch_type cannot be empty")
	}

	// 1. Fetch current game state
	state, err := s.GetGameState(sessionID, puzzle)
	if err != nil {
		return nil, err
	}

	if state.Status != "active" {
		return nil, fmt.Errorf("game already completed: status is %s", state.Status)
	}

	if state.PitchGuessed {
		return nil, fmt.Errorf("pitch type was already guessed in this session")
	}

	// 2. Fetch target pitch profile
	targetPitchProfile, err := s.repo.GetPitchProfileByID(puzzle.TargetPitchProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target pitch profile: %w", err)
	}

	matched := strings.EqualFold(pitchType, targetPitchProfile.PitchType)

	// 3. Save pitch guess
	newPitchGuess := &models.PitchGuess{
		SessionID:        sessionID,
		PuzzleID:         puzzle.ID,
		GuessedPitchType: pitchType,
		Matched:          matched,
	}
	err = s.repo.SavePitchGuess(newPitchGuess)
	if err != nil {
		return nil, fmt.Errorf("failed to save pitch guess: %w", err)
	}

	// 4. Update game state
	state.PitchGuessed = true
	state.PitchGuess = &PitchFeedback{
		GuessedType: pitchType,
		ActualType:  targetPitchProfile.PitchType,
		Matched:     matched,
		Velocity:    targetPitchProfile.Velocity,
		SpinRate:    targetPitchProfile.SpinRate,
	}

	return state, nil
}
