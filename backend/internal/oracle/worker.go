// Package oracle Oracle Worker 自动结算服务
package oracle

// 导入依赖
import (
	"context" // 上下文
	"log"     // 日志
	"strconv" // 数字转字符串
	"time"    // 时间

	"github.com/prediction-did/simple/internal/alert"      // 告警
	"github.com/prediction-did/simple/internal/blockchain" // 链上交互
	"github.com/prediction-did/simple/internal/config"     // 配置
	"github.com/prediction-did/simple/internal/repository" // 仓储
	"github.com/prediction-did/simple/internal/wcprovider" // 比分数据源
)

// Worker Oracle 工作者：轮询并处理到期结算任务
type Worker struct {
	cfg     *config.Config            // 配置
	jobs    *repository.OracleJobRepo // 任务仓储
	matches *repository.MatchRepo     // 比赛仓储
	markets *repository.MarketRepo    // 市场仓储
	dual    *wcprovider.DualProvider  // 双数据源提供者
	chain   *blockchain.OracleClient  // 链上 Oracle 客户端
	alerts  *alert.Notifier           // 告警通知
}

// NewWorker 创建 Worker 并注入所有依赖
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

// Run 启动 Worker 主循环
func (w *Worker) Run(ctx context.Context) error {
	// 按配置间隔轮询
	ticker := time.NewTicker(time.Duration(w.cfg.OraclePollSeconds) * time.Second)
	defer ticker.Stop()
	for {
		// 执行一次 tick
		if err := w.tick(ctx); err != nil {
			log.Printf("oracle tick: %v", err)
		}
		select {
		case <-ctx.Done(): // 退出
			return ctx.Err()
		case <-ticker.C: // 等待下一次
		}
	}
}

// tick 单次调度：入队新任务 + 处理到期任务
func (w *Worker) tick(ctx context.Context) error {
	// 将 FINISHED 的比赛入队
	if err := w.enqueueFinished(ctx); err != nil {
		return err
	}
	// 处理到期任务
	return w.processDue(ctx)
}

// enqueueFinished 查找 FINISHED 的比赛，为其关联市场创建 Oracle 任务
func (w *Worker) enqueueFinished(ctx context.Context) error {
	// 查询所有 FINISHED 状态的比赛
	matches, err := w.matches.List(ctx, "FINISHED", 100, 0)
	if err != nil {
		return err
	}
	// 冷却时间
	cooldown := time.Duration(w.cfg.OracleCooldownMinutes) * time.Minute
	for _, m := range matches {
		// 查找该比赛对应的 OPEN 市场
		markets, err := w.markets.ListOpenByMatchID(ctx, m.ID)
		if err != nil {
			return err
		}
		for _, mk := range markets {
			// 检查是否已有活跃任务
			active, err := w.jobs.HasActiveForMarket(ctx, mk.ID)
			if err != nil || active {
				continue // 已有则跳过
			}
			// 创建任务，冷却后执行
			executeAfter := time.Now().Add(cooldown)
			if _, err := w.jobs.Create(ctx, mk.ID, &m.ID, executeAfter); err != nil {
				log.Printf("oracle enqueue: %v", err)
			}
			// 更新比赛状态
			_ = w.matches.SetStatus(ctx, m.ID, "ORACLE_PENDING")
		}
	}
	return nil
}

// processDue 处理所有到期的任务
func (w *Worker) processDue(ctx context.Context) error {
	// 未配置链客户端则不处理
	if w.chain == nil {
		return nil
	}
	// 获取已到期任务
	due, err := w.jobs.ListDue(ctx, time.Now())
	if err != nil {
		return err
	}
	// 逐个处理
	for _, job := range due {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("oracle job %d: %v", job.ID, err)
		}
	}
	return nil
}

// processJob 处理单个结算任务
func (w *Worker) processJob(ctx context.Context, job repository.OracleJob) error {
	// 无关联比赛则跳过
	if job.MatchID == nil {
		return nil
	}
	// 获取比赛详情
	match, err := w.matches.GetByID(ctx, *job.MatchID)
	if err != nil {
		return err
	}
	// 从双源获取比分并比对
	pri, sec, ok, err := w.dual.CompareScores(match.ExternalID)
	if err != nil {
		return err
	}
	// 构造元数据字段
	fields := map[string]interface{}{
		"primary_home":   pri.HomeScore, // 主源主队得分
		"primary_away":   pri.AwayScore, // 主源客队得分
		"secondary_home": sec.HomeScore, // 备源主队得分
		"secondary_away": sec.AwayScore, // 备源客队得分
	}
	// 双源不一致时进入人工审核
	if !ok {
		_ = w.jobs.UpdateStatus(ctx, job.ID, "manual_review", merge(fields, map[string]interface{}{
			"error_message": "dual source score mismatch or missing",
		}))
		w.alerts.Send("oracle_manual_review", "job "+strconv.FormatInt(job.ID, 10)+" needs review")
		return nil
	}
	// 获取市场信息
	mk, err := w.markets.GetByID(ctx, job.MarketID)
	if err != nil {
		return err
	}
	// 确定结算规则
	rule := "HOME_WIN"
	if mk.ResolutionRule != "" {
		rule = mk.ResolutionRule
	}
	// 根据规则与比分计算结果
	outcome := wcprovider.OutcomeFromRule(rule, pri.HomeScore, pri.AwayScore)
	fields["proposed_outcome"] = outcome

	// 发起链上结算交易
	var txHash string
	if w.cfg.OracleTimelockSeconds <= 0 {
		// 无时间锁直接结算
		txHash, err = w.chain.ResolveNow(ctx, job.MarketAddress, uint8(outcome))
	} else {
		// 有时间锁：先请求 -> 等待确认
		txHash, err = w.chain.RequestResolve(ctx, job.MarketAddress, uint8(outcome))
		if err == nil {
			_ = w.chain.WaitMined(ctx, txHash)
			// 等待时间锁到期
			time.Sleep(time.Duration(w.cfg.OracleTimelockSeconds) * time.Second)
			// 确认结算
			txHash, err = w.chain.ConfirmResolve(ctx, job.MarketAddress)
		}
	}
	// 链上失败
	if err != nil {
		_ = w.jobs.UpdateStatus(ctx, job.ID, "failed", merge(fields, map[string]interface{}{
			"error_message": err.Error(),
		}))
		w.alerts.Send("oracle_failed", err.Error())
		return err
	}
	// 等待上链
	_ = w.chain.WaitMined(ctx, txHash)
	// 更新任务状态为已确认
	_ = w.jobs.UpdateStatus(ctx, job.ID, "confirmed", merge(fields, map[string]interface{}{
		"tx_hash": txHash,
	}))
	// 更新市场为已结算
	_ = w.markets.UpdateResolved(ctx, job.MarketAddress, outcome, "0", "0")
	// 更新比赛状态
	if job.MatchID != nil {
		_ = w.matches.SetStatus(ctx, *job.MatchID, "RESOLVED")
	}
	return nil
}

// merge 合并两个 map
func merge(base map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	for k, v := range extra {
		base[k] = v
	}
	return base
}
