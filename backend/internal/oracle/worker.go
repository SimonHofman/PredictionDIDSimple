package oracle

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/prediction-did/simple/internal/alert"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/wcprovider"
)

type Worker struct {
	cfg      *config.Config
	jobs     *repository.OracleJobRepo
	matches  *repository.MatchRepo
	markets  *repository.MarketRepo
	dual     *wcprovider.DualProvider
	chain    *blockchain.OracleClient
	alerts   *alert.Notifier
}

func NewWorker(
	cfg *config.Config,
	jobs *repository.OracleJobRepo,
	matches *repository.MatchRepo,
	markets *repository.MarketRepo,
	dual *wcprovider.DualProvider,
	chain *blockchain.OracleClient,
	alerts *alert.Notifier,
) *Worker {
	return &Worker{cfg: cfg, jobs: jobs, matches: matches, markets: markets, dual: dual, chain: chain, alerts: alerts}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(w.cfg.OraclePollSeconds) * time.Second)
	defer ticker.Stop()
	for {
		if err := w.tick(ctx); err != nil {
			log.Printf("oracle tick: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) tick(ctx context.Context) error {
	if err := w.enqueueFinished(ctx); err != nil {
		return err
	}
	return w.processDue(ctx)
}

func (w *Worker) enqueueFinished(ctx context.Context) error {
	matches, err := w.matches.List(ctx, "FINISHED", 100, 0)
	if err != nil {
		return err
	}
	cooldown := time.Duration(w.cfg.OracleCooldownMinutes) * time.Minute
	for _, m := range matches {
		markets, err := w.markets.ListOpenByMatchID(ctx, m.ID)
		if err != nil {
			return err
		}
		for _, mk := range markets {
			active, err := w.jobs.HasActiveForMarket(ctx, mk.ID)
			if err != nil || active {
				continue
			}
			executeAfter := time.Now().Add(cooldown)
			if _, err := w.jobs.Create(ctx, mk.ID, &m.ID, executeAfter); err != nil {
				log.Printf("oracle enqueue: %v", err)
			}
			_ = w.matches.SetStatus(ctx, m.ID, "ORACLE_PENDING")
		}
	}
	return nil
}

func (w *Worker) processDue(ctx context.Context) error {
	if w.chain == nil {
		return nil
	}
	due, err := w.jobs.ListDue(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, job := range due {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("oracle job %d: %v", job.ID, err)
		}
	}
	return nil
}

func (w *Worker) processJob(ctx context.Context, job repository.OracleJob) error {
	if job.MatchID == nil {
		return nil
	}
	match, err := w.matches.GetByID(ctx, *job.MatchID)
	if err != nil {
		return err
	}
	pri, sec, ok, err := w.dual.CompareScores(match.ExternalID)
	if err != nil {
		return err
	}
	fields := map[string]interface{}{
		"primary_home":   pri.HomeScore,
		"primary_away":   pri.AwayScore,
		"secondary_home": sec.HomeScore,
		"secondary_away": sec.AwayScore,
	}
	if !ok {
		_ = w.jobs.UpdateStatus(ctx, job.ID, "manual_review", merge(fields, map[string]interface{}{
			"error_message": "dual source score mismatch or missing",
		}))
		w.alerts.Send("oracle_manual_review", "job "+strconv.FormatInt(job.ID, 10)+" needs review")
		return nil
	}
	mk, err := w.markets.GetByID(ctx, job.MarketID)
	if err != nil {
		return err
	}
	rule := "HOME_WIN"
	if mk.ResolutionRule != "" {
		rule = mk.ResolutionRule
	}
	outcome := wcprovider.OutcomeFromRule(rule, pri.HomeScore, pri.AwayScore)
	fields["proposed_outcome"] = outcome

	var txHash string
	if w.cfg.OracleTimelockSeconds <= 0 {
		txHash, err = w.chain.ResolveNow(ctx, job.MarketAddress, uint8(outcome))
	} else {
		txHash, err = w.chain.RequestResolve(ctx, job.MarketAddress, uint8(outcome))
		if err == nil {
			_ = w.chain.WaitMined(ctx, txHash)
			time.Sleep(time.Duration(w.cfg.OracleTimelockSeconds) * time.Second)
			txHash, err = w.chain.ConfirmResolve(ctx, job.MarketAddress)
		}
	}
	if err != nil {
		_ = w.jobs.UpdateStatus(ctx, job.ID, "failed", merge(fields, map[string]interface{}{
			"error_message": err.Error(),
		}))
		w.alerts.Send("oracle_failed", err.Error())
		return err
	}
	_ = w.chain.WaitMined(ctx, txHash)
	_ = w.jobs.UpdateStatus(ctx, job.ID, "confirmed", merge(fields, map[string]interface{}{
		"tx_hash": txHash,
	}))
	_ = w.markets.UpdateResolved(ctx, job.MarketAddress, outcome, "0", "0")
	if job.MatchID != nil {
		_ = w.matches.SetStatus(ctx, *job.MatchID, "RESOLVED")
	}
	return nil
}

func merge(base map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	for k, v := range extra {
		base[k] = v
	}
	return base
}
