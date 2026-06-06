// Package wcprovider 提供世界杯/体育赛事数据的双源供应和同步功能
package wcprovider

// 导入依赖
import (
	"context"       // 上下文，用于控制请求生命周期
	"encoding/json" // JSON 序列化和反序列化
	"os"            // 操作系统相关，用于文件读取

	"github.com/prediction-did/simple/internal/models"     // 数据模型定义
	"github.com/prediction-did/simple/internal/repository" // 数据仓储层
)

// ScoreSnapshot 比分快照结构体
// 记录某场比赛在某一时刻的比分状态
type ScoreSnapshot struct {
	ExternalID string // 外部数据源的比赛唯一标识
	HomeScore  int    // 主队得分
	AwayScore  int    // 客队得分
	Status     string // 比赛状态（如 "FINISHED"、"LIVE" 等）
}

// DualProvider 双数据源供应商结构体
// 支持从两个独立数据源获取比赛数据，用于交叉验证
type DualProvider struct {
	primaryPath   string // 主数据源文件路径
	secondaryPath string // 备用数据源文件路径
}

// NewDual 创建新的双数据源供应商实例
// 参数 primary: 主数据源文件路径
// 参数 secondary: 备用数据源文件路径
// 返回: DualProvider 指针
func NewDual(primary, secondary string) *DualProvider {
	return &DualProvider{primaryPath: primary, secondaryPath: secondary} // 初始化双源供应商
}

// load 从指定 JSON 文件加载比赛比分数据
// 只加载有完整比分数据的比赛（HomeScore 和 AwayScore 都不为 nil）
// 参数 path: JSON 数据文件路径
// 返回: 以外部ID为键的比分快照映射和错误信息
func (p *DualProvider) load(path string) (map[string]ScoreSnapshot, error) {
	raw, err := os.ReadFile(path) // 读取文件内容
	if err != nil {
		return nil, err // 文件读取失败
	}
	var items []mockMatch // 解析 JSON 数组
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err // JSON 解析失败
	}
	out := make(map[string]ScoreSnapshot) // 初始化结果映射
	for _, it := range items {            // 遍历所有比赛数据
		if it.HomeScore == nil || it.AwayScore == nil {
			continue // 跳过没有完整比分的比赛
		}
		// 构建比分快照并以外部ID为键存储
		out[it.ExternalID] = ScoreSnapshot{
			ExternalID: it.ExternalID, // 外部ID
			HomeScore:  *it.HomeScore, // 主队得分
			AwayScore:  *it.AwayScore, // 客队得分
			Status:     it.Status,     // 比赛状态
		}
	}
	return out, nil // 返回比分快照映射
}

// CompareScores 比较两个数据源中指定比赛的比分是否一致
// 用于数据交叉验证，确保预言机数据的可靠性
// 参数 externalID: 外部比赛唯一标识
// 返回: 主源比分、备源比分、是否匹配、错误信息
func (p *DualProvider) CompareScores(externalID string) (primary, secondary ScoreSnapshot, match bool, err error) {
	pri, err := p.load(p.primaryPath) // 加载主数据源
	if err != nil {
		return primary, secondary, false, err // 主源加载失败
	}
	sec, err := p.load(p.secondaryPath) // 加载备用数据源
	if err != nil {
		return primary, secondary, false, err // 备源加载失败
	}
	var ok1, ok2 bool                // 标记两个源是否都包含该比赛
	primary, ok1 = pri[externalID]   // 从主源获取比分
	secondary, ok2 = sec[externalID] // 从备源获取比分
	if !ok1 || !ok2 {
		return primary, secondary, false, nil // 有一个源缺少数据
	}
	// 比较主客场比分是否完全一致
	match = primary.HomeScore == secondary.HomeScore && primary.AwayScore == secondary.AwayScore
	return primary, secondary, match, nil // 返回比较结果
}

// SyncPrimary 使用主数据源同步比赛数据到数据库
// 参数 ctx: 上下文
// 参数 repo: 比赛仓储
// 返回: 同步的记录数和错误信息
func (p *DualProvider) SyncPrimary(ctx context.Context, repo *repository.MatchRepo) (int, error) {
	return NewMock(p.primaryPath).Sync(ctx, repo) // 委托给 MockProvider 执行同步
}

// OutcomeFromRule 根据预测规则和比分计算比赛结果
// 返回 0 表示 Yes（条件满足），1 表示 No（条件不满足）
// 参数 rule: 预测规则名称（如 "OVER_25"、"HOME_WIN"）
// 参数 home: 主队得分
// 参数 away: 客队得分
// 返回: 结果值（0=Yes, 1=No）
func OutcomeFromRule(rule string, home, away int) int {
	switch rule {
	case "OVER_25":
		// 总进球数大于 2.5 规则
		if home+away > 2 {
			return 0 // Yes：总进球超过 2.5
		}
		return 1 // No：总进球不超过 2.5
	default: // HOME_WIN 主队获胜规则（默认）
		if home > away {
			return 0 // Yes：主队获胜
		}
		return 1 // No：主队未获胜
	}
}

// FinishMatch 完成比赛并更新最终比分
// 将比赛状态标记为 "FINISHED" 并记录最终比分
// 参数 ctx: 上下文
// 参数 repo: 比赛仓储
// 参数 externalID: 外部比赛唯一标识
// 参数 home: 主队最终得分
// 参数 away: 客队最终得分
// 返回: 错误信息
func (p *DualProvider) FinishMatch(ctx context.Context, repo *repository.MatchRepo, externalID string, home, away int) error {
	m, err := repo.GetByExternalID(ctx, externalID) // 通过外部ID获取比赛记录
	if err != nil {
		return err // 查询失败
	}
	m.Status = "FINISHED"                          // 设置比赛状态为已结束
	m.HomeScore = &home                            // 设置主队最终得分
	m.AwayScore = &away                            // 设置客队最终得分
	out := OutcomeFromRule("HOME_WIN", home, away) // 计算比赛结果
	_ = out                                        // 暂未使用计算结果（保留供后续扩展）
	return repo.Upsert(ctx, *m)                    // 更新比赛记录到数据库
}

// 确保 models.Match 类型被使用（编译时检查）
var _ = models.Match{}
