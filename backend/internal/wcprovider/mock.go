package wcprovider

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/prediction-did/simple/internal/models"
	"github.com/prediction-did/simple/internal/repository"
)

type MockProvider struct {
	path string
}

func NewMock(path string) *MockProvider {
	return &MockProvider{path: path}
}

type mockMatch struct {
	ExternalID string `json:"external_id"`
	HomeTeam   string `json:"home_team"`
	AwayTeam   string `json:"away_team"`
	KickoffAt  string `json:"kickoff_at"`
	Status     string `json:"status"`
	HomeScore  *int   `json:"home_score"`
	AwayScore  *int   `json:"away_score"`
}

func (p *MockProvider) Sync(ctx context.Context, repo *repository.MatchRepo) (int, error) {
	raw, err := os.ReadFile(p.path)
	if err != nil {
		return 0, err
	}
	var items []mockMatch
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0, err
	}
	n := 0
	for _, it := range items {
		kick, err := time.Parse(time.RFC3339, it.KickoffAt)
		if err != nil {
			return n, err
		}
		m := models.Match{
			ExternalID: it.ExternalID,
			HomeTeam:   it.HomeTeam,
			AwayTeam:   it.AwayTeam,
			KickoffAt:  kick,
			Status:     it.Status,
			HomeScore:  it.HomeScore,
			AwayScore:  it.AwayScore,
		}
		if err := repo.Upsert(ctx, m); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
