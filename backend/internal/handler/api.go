// Package handler API 路由注册与通用工具函数
package handler

// 导入依赖
import (
	"encoding/json" // JSON 编码
	"net/http"      // HTTP
	"strconv"       // 字符串转数字

	"github.com/go-chi/chi/v5"                             // Chi 路由
	"github.com/prediction-did/simple/internal/auth"       // 认证中间件
	"github.com/prediction-did/simple/internal/blockchain" // 区块链客户端
	"github.com/prediction-did/simple/internal/config"     // 配置
	"github.com/prediction-did/simple/internal/repository" // 仓储层
	"github.com/prediction-did/simple/internal/vc"         // 可验证凭证
)

// API 聚合所有 handler 所需的依赖
type API struct {
	Cfg         *config.Config             // 全局配置
	Matches     *repository.MatchRepo      // 比赛仓储
	Markets     *repository.MarketRepo     // 市场仓储
	Users       *repository.UserRepo       // 用户仓储
	Positions   *repository.PositionRepo   // 持仓仓储
	OracleJobs  *repository.OracleJobRepo  // Oracle 任务仓储
	Credentials *repository.CredentialRepo // 凭证仓储
	Stats       *repository.StatsRepo      // 统计仓储
	Compliance  *repository.ComplianceRepo // 合规仓储
	VCIssuer    *vc.Issuer                 // VC 颁发器
	OracleChain *blockchain.OracleClient   // Oracle 链上客户端
}

// RegisterRoutes 注册所有路由到 Chi Router
func (a *API) RegisterRoutes(r chi.Router) {
	// ========== 公开接口 ==========
	r.Get("/matches", a.listMatches)                        // 比赛列表
	r.Get("/matches/{id}", a.getMatch)                      // 比赛详情
	r.Get("/markets", a.listMarkets)                        // 市场列表
	r.Get("/markets/{id}", a.getMarket)                     // 市场详情
	r.Get("/markets/{id}/pool", a.marketPool)               // 市场资金池
	r.Get("/markets/{id}/orderbook", a.marketOrderbook)     // 市场订单簿
	r.Get("/stats/platform", a.platformStats)               // 平台统计
	r.Get("/compliance/restricted", a.complianceRestricted) // 合规受限名单
	r.Post("/kyc/webhook", a.kycWebhook)                    // KYC 回调
	r.Get("/events/scores", a.streamScores)                 // SSE 比分流
	r.Get("/metrics", a.prometheusMetrics)                  // Prometheus 指标

	// ========== 登录接口 ==========
	r.Post("/auth/siwe", a.siweAuth)      // SIWE 钱包登录
	r.Post("/auth/verify-vc", a.verifyVC) // 验证 VC

	// ========== 需要 JWT 认证的接口 ==========
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(a.Cfg.JWTSecret))         // JWT 中间件
		r.Get("/me/positions", a.myPositions)           // 我的持仓
		r.Post("/users/bind-did", a.bindDID)            // 绑定 DID
		r.Get("/users/me/credentials", a.myCredentials) // 我的凭证
	})

	// ========== 需要管理员 Key 的接口 ==========
	r.Group(func(r chi.Router) {
		r.Use(auth.AdminMiddleware(a.Cfg.AdminAPIKey))            // 管理员中间件
		r.Get("/admin/oracle-jobs", a.listOracleJobs)             // Oracle 任务列表
		r.Post("/admin/markets", a.registerMarket)                // 注册市场
		r.Post("/admin/markets/{id}/void", a.voidMarket)          // 作废市场
		r.Post("/admin/oracle-jobs/{id}/retry", a.retryOracleJob) // 重试任务
		r.Post("/credentials/issue", a.issueCredential)           // 颁发凭证
	})
}

// writeJSON 通用 JSON 响应写入
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json") // 设置 Content-Type
	w.WriteHeader(status)                              // 写入状态码
	_ = json.NewEncoder(w).Encode(v)                   // 序列化写入
}

// writeError 写入 JSON 错误响应
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// queryInt 从请求 query 参数取整数，失败返回默认值
func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key) // 取值
	if v == "" {
		return def // 空则默认
	}
	n, err := strconv.Atoi(v) // 转整数
	if err != nil {
		return def
	}
	return n
}

// parseID 从 URL 路径参数解析 int64 ID
func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
