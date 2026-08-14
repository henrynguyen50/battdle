package normalizer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pitchle/shared/models"
)

var MLBTeamIDs = []int{
	108, 109, 110, 111, 112, 113, 114, 115, 116, 117,
	118, 119, 120, 121, 133, 134, 135, 136, 137, 138,
	139, 140, 141, 142, 143, 144, 145, 146, 147, 158,
}

func NormalizeName(name string) string {
	parts := strings.Split(name, ",")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(name)
}

type PeopleResponse struct {
	People []struct {
		ID              int    `json:"id"`
		FullName        string `json:"fullName"`
		BirthDate       string `json:"birthDate"`
		MLBDebutDate    string `json:"mlbDebutDate"`
		LastPlayedDate  string `json:"lastPlayedDate"`
		Active          bool   `json:"active"`
		PrimaryPosition struct {
			Abbreviation string `json:"abbreviation"`
		} `json:"primaryPosition"`
		CurrentTeam *struct {
			ID int `json:"id"`
		} `json:"currentTeam"`
	} `json:"people"`
}

func chunkSlice(slice []int, chunkSize int) [][]int {
	var chunks [][]int
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// FetchAllPlayerMetadata pre-fetches player details from the MLB Stats API in bulk chunks.
func FetchAllPlayerMetadata(playerIDs []int, teamMap map[int]int) (map[int]models.Player, error) {
	cache := make(map[int]models.Player)
	chunks := chunkSlice(playerIDs, 100)
	client := &http.Client{Timeout: 5 * time.Second}

	for _, chunk := range chunks {
		var idStrs []string
		for _, id := range chunk {
			idStrs = append(idStrs, strconv.Itoa(id))
		}
		personIDsParam := strings.Join(idStrs, ",")

		url := fmt.Sprintf("https://statsapi.mlb.com/api/v1/people?personIds=%s&hydrate=currentTeam", personIDsParam)
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("http request failed for chunk: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var data PeopleResponse
		err = json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("json decode failed: %w", err)
		}

		for _, p := range data.People {
			// Parse birth_date
			var birthDate *time.Time
			var birthYear int
			if p.BirthDate != "" {
				t, err := time.Parse("2006-01-02", p.BirthDate)
				if err == nil {
					birthDate = &t
					birthYear = t.Year()
				}
			}
			if birthYear == 0 {
				birthYear = 1990 // fallback
			}

			// Parse debut year
			var debutYear int
			if p.MLBDebutDate != "" {
				t, err := time.Parse("2006-01-02", p.MLBDebutDate)
				if err == nil {
					debutYear = t.Year()
				}
			}
			if debutYear == 0 {
				debutYear = birthYear + 22
			}

			// Parse last year
			lastYear := 2026
			if !p.Active && p.LastPlayedDate != "" {
				t, err := time.Parse("2006-01-02", p.LastPlayedDate)
				if err == nil {
					lastYear = t.Year()
				}
			}

			// Parse position
			pos := p.PrimaryPosition.Abbreviation
			if pos == "" {
				pos = "P"
			}

			// Parse team ID
			var dbTeamID int
			if p.CurrentTeam != nil {
				if id, ok := teamMap[p.CurrentTeam.ID]; ok {
					dbTeamID = id
				}
			}

			// Fallback team mapping if not found in db
			if dbTeamID == 0 {
				mlbTeamID := MLBTeamIDs[p.ID%len(MLBTeamIDs)]
				dbTeamID = teamMap[mlbTeamID]
			}

			cache[p.ID] = models.Player{
				MLBID:        p.ID,
				BirthDate:    birthDate,
				BirthYear:    birthYear,
				Position:     pos,
				MLBDebutYear: debutYear,
				MLBLastYear:  lastYear,
				TeamID:       dbTeamID,
			}
		}
	}

	return cache, nil
}

// GenerateMetadata looks up from bulk cached player info, falling back to deterministic generation on miss.
func GenerateMetadata(playerID int, teamMap map[int]int, cache map[int]models.Player) (models.Player, error) {
	if cache != nil {
		if p, ok := cache[playerID]; ok {
			return p, nil
		}
	}
	return GenerateDeterministicMetadata(playerID, teamMap)
}

// GenerateDeterministicMetadata generates deterministic metadata based on playerID.
func GenerateDeterministicMetadata(playerID int, teamMap map[int]int) (models.Player, error) {
	birthYear := 1985 + (playerID % 18)
	birthDateStr := fmt.Sprintf("%d-06-15", birthYear)
	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		return models.Player{}, fmt.Errorf("failed to parse birth date: %w", err)
	}

	debutYear := birthYear + 20 + (playerID % 5)
	lastYear := debutYear + (playerID % 8)
	if lastYear > 2026 {
		lastYear = 2026
	}
	if lastYear < debutYear {
		lastYear = debutYear
	}

	mlbTeamID := MLBTeamIDs[playerID%len(MLBTeamIDs)]
	dbTeamID, ok := teamMap[mlbTeamID]
	if !ok {
		return models.Player{}, fmt.Errorf("MLB team ID %d not found in team map", mlbTeamID)
	}

	return models.Player{
		MLBID:        playerID,
		BirthDate:    &birthDate,
		BirthYear:    birthYear,
		Position:     "P",
		MLBDebutYear: debutYear,
		MLBLastYear:  lastYear,
		TeamID:       dbTeamID,
	}, nil
}
