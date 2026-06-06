// Package wcprovider 提供世界杯/体育赛事数据的双源供应和同步功能
package wcprovider

// 导入依赖
import (
	"context"       // 上下文，用于控制请求生命周期
	"encoding/json" // JSON 序列化和反序列化
	"os"            // 操作系统相关，用于文件读取
	"time"          // 时间处理

	"github.com/prediction-did/simple/internal/models"     // 数据模型定义
	"github.com/prediction-did/simple/internal/repository" // 数据仓储层
)

// MockProvider 模拟数据供应商结构体
// 从本地 JSON 文件加载比赛数据，用于开发和测试环境
type MockProvider struct {
	path string // 模拟数据 JSON 文件路径
}

// NewMock 创建新的模拟数据供应商实例
// 参数 path: 模拟数据 JSON 文件路径
// 返回: MockProvider 指针
func NewMock(path string) *MockProvider {
	return &MockProvider{path: path} // 初始化并返回模拟供应商
}

// mockMatch 模拟比赛数据的 JSON 映射结构体
// 对应本地 JSON 文件中的比赛记录格式
type mockMatch struct {
	ExternalID string `json:"external_id"` // 外部数据源比赛唯一标识
	HomeTeam   string `json:"home_team"`   // 主队名称
	AwayTeam   string `json:"away_team"`   // 客队名称
	KickoffAt  string `json:"kickoff_at"`  // 开球时间（RFC3339 格式字符串）
	Status     string `json:"status"`      // 比赛状态
	HomeScore  *int   `json:"home_score"`  // 主队得分（可空，比赛未结束时为 nil）
	AwayScore  *int   `json:"away_score"`  // 客队得分（可空，比赛未结束时为 nil）
}

// Sync 将模拟数据文件中的比赛记录同步到数据库
// 读取 JSON 文件，解析每条比赛记录，并逐一通过仓储层写入数据库
// 参数 ctx: 上下文
// 参数 repo: 比赛仓储，用于执行数据库操作
// 返回: 成功同步的记录数和错误信息
func (p *MockProvider) Sync(ctx context.Context, repo *repository.MatchRepo) (int, error) {
	raw, err := os.ReadFile(p.path) // 读取模拟数据文件
	if err != nil {
		return 0, err // 文件读取失败
	}
	var items []mockMatch // 存储解析后的比赛数据数组
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0, err // JSON 解析失败
	}
	n := 0                     // 初始化成功计数器
	for _, it := range items { // 遍历每条比赛记录
		kick, err := time.Parse(time.RFC3339, it.KickoffAt) // 解析开球时间字符串为 time.Time
		if err != nil {
			return n, err // 时间解析失败，返回已同步的数量
		}
		// 构建比赛模型对象
		m := models.Match{
			ExternalID: it.ExternalID, // 外部ID
			HomeTeam:   it.HomeTeam,   // 主队
			AwayTeam:   it.AwayTeam,   // 客队
			KickoffAt:  kick,          // 开球时间
			Status:     it.Status,     // 状态
			HomeScore:  it.HomeScore,  // 主队得分
			AwayScore:  it.AwayScore,  // 客队得分
		}
		if err := repo.Upsert(ctx, m); err != nil {
			return n, err // 数据库写入失败，返回已同步数量
		}
		n++ // 成功计数加一
	}
	return n, nil // 返回成功同步的总记录数
}
