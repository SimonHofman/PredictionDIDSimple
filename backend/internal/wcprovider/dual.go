package wcprovider

import (
	"context"
	"encoding/json"
	"os"

	"github.com/prediction-did/simple/internal/models"
	"github.com/prediction-did/simple/internal/repository"
)

type ScoreSnapshot struct {
	ExternalID string
	HomeScore  int
	AwayScore  int
	Status     string
}

type DualProvider struct {
	primaryPath   string
	secondaryPath string
}

func NewDual(primary, secondary string) *DualProvider {
	return &DualProvider{primaryPath: primary, secondaryPath: secondary}
}

func (p *DualProvider) load(path string) (map[string]ScoreSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []mockMatch
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	out := make(map[string]ScoreSnapshot)
	for _, it := range items {
		if it.HomeScore == nil || it.AwayScore == nil {
			continue
		}
		out[it.ExternalID] = ScoreSnapshot{
			ExternalID: it.ExternalID,
			HomeScore:  *it.HomeScore,
			AwayScore:  *it.AwayScore,
			Status:     it.Status,
		}
	}
	return out, nil
}

func (p *DualProvider) CompareScores(externalID string) (primary, secondary ScoreSnapshot, match bool, err error) {
	pri, err := p.load(p.primaryPath)
	if err != nil {
		return primary, secondary, false, err
	}
	sec, err := p.load(p.secondaryPath)
	if err != nil {
		return primary, secondary, false, err
	}
	var ok1, ok2 bool
	primary, ok1 = pri[externalID]
	secondary, ok2 = sec[externalID]
	if !ok1 || !ok2 {
		return primary, secondary, false, nil
	}
	match = primary.HomeScore == secondary.HomeScore && primary.AwayScore == secondary.AwayScore
	return primary, secondary, match, nil
}

func (p *DualProvider) SyncPrimary(ctx context.Context, repo *repository.MatchRepo) (int, error) {
	return NewMock(p.primaryPath).Sync(ctx, repo)
}

func OutcomeFromRule(rule string, home, away int) int {
	switch rule {
	case "OVER_25":
		if home+away > 2 {
			return 0
		}
		return 1
	default: // HOME_WIN
		if home > away {
			return 0
		}
		return 1
	}
}

func (p *DualProvider) FinishMatch(ctx context.Context, repo *repository.MatchRepo, externalID string, home, away int) error {
	m, err := repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	m.Status = "FINISHED"
	m.HomeScore = &home
	m.AwayScore = &away
	out := OutcomeFromRule("HOME_WIN", home, away)
	_ = out
	return repo.Upsert(ctx, *m)
}

var _ = models.Match{}
