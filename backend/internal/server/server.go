// Package server 提供 HTTP 服务器的初始化和生命周期管理
package server

// 导入依赖
import (
	"context"  // 上下文，用于服务器优雅关闭
	"fmt"      // 格式化输出
	"net/http" // HTTP 服务器标准库
	"time"     // 时间处理

	"github.com/go-chi/chi/v5"                                   // Chi 路由框架
	"github.com/go-chi/chi/v5/middleware"                        // Chi 内置中间件
	"github.com/go-chi/cors"                                     // CORS 跨域资源共享中间件
	"github.com/jackc/pgx/v5/pgxpool"                            // PostgreSQL 连接池
	"github.com/prediction-did/simple/internal/blockchain"       // 区块链客户端
	"github.com/prediction-did/simple/internal/config"           // 配置管理
	"github.com/prediction-did/simple/internal/handler"          // HTTP 请求处理器
	appmw "github.com/prediction-did/simple/internal/middleware" // 应用自定义中间件（别名 appmw）
	"github.com/prediction-did/simple/internal/repository"       // 数据仓储层
	"github.com/prediction-did/simple/internal/vc"               // 可验证凭证模块
	"github.com/redis/go-redis/v9"                               // Redis 客户端
)

// Server HTTP 服务器结构体，封装底层 http.Server
type Server struct {
	httpServer *http.Server // 底层 HTTP 服务器实例
}

// Dependencies 服务器依赖注入结构体
// 包含服务器运行所需的所有外部依赖
type Dependencies struct {
	Port        string                   // 服务器监听端口
	Cfg         *config.Config           // 应用配置
	DB          *pgxpool.Pool            // 数据库连接池
	Redis       *redis.Client            // Redis 客户端
	Chain       *blockchain.Client       // 区块链客户端（通用）
	OracleChain *blockchain.OracleClient // 预言机区块链客户端
}

// New 创建并配置新的 HTTP 服务器实例
// 初始化路由、中间件、处理器，并注册所有 API 路由
// 参数 deps: 服务器所有外部依赖
// 返回: 配置完成的 Server 实例
func New(deps Dependencies) *Server {
	r := chi.NewRouter() // 创建新的 Chi 路由器

	// 注册全局中间件
	r.Use(middleware.RequestID)                         // 为每个请求分配唯一 ID
	r.Use(middleware.RealIP)                            // 获取真实客户端 IP（支持代理）
	r.Use(middleware.Logger)                            // 请求日志记录
	r.Use(middleware.Recoverer)                         // 从 panic 中恢复，防止服务器崩溃
	r.Use(appmw.RateLimit(deps.Cfg.RateLimitPerMinute)) // 速率限制中间件

	// 初始化合规仓储，用于地理封锁日志记录
	compRepo := repository.NewComplianceRepo(deps.DB)
	// 注册地理封锁中间件，记录每次访问的地理信息
	r.Use(appmw.GeoBlock(deps.Cfg, func(ip, country, path string, allowed bool) {
		_ = compRepo.LogGeo(context.Background(), ip, country, path, allowed) // 异步记录地理访问日志
	}))

	// 配置 CORS 跨域策略
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},                                                              // 允许的前端来源
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                                                                     // 允许的 HTTP 方法
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Admin-Key", "CF-IPCountry", "X-Country-Code", "X-KYC-Signature"}, // 允许的请求头
		AllowCredentials: true,                                                                                                                    // 允许携带凭证（Cookie 等）
	}))

	// 初始化健康检查处理器并注册路由
	health := &handler.Health{
		DB:    deps.DB,    // 数据库连接池
		Redis: deps.Redis, // Redis 客户端
		Chain: deps.Chain, // 区块链客户端
	}
	health.RegisterRoutes(r) // 注册健康检查路由

	// 初始化 API 处理器，注入所有依赖的仓储和服务
	api := &handler.API{
		Cfg:         deps.Cfg,                              // 应用配置
		Matches:     repository.NewMatchRepo(deps.DB),      // 比赛仓储
		Markets:     repository.NewMarketRepo(deps.DB),     // 市场仓储
		Users:       repository.NewUserRepo(deps.DB),       // 用户仓储
		Positions:   repository.NewPositionRepo(deps.DB),   // 持仓仓储
		OracleJobs:  repository.NewOracleJobRepo(deps.DB),  // 预言机任务仓储
		Credentials: repository.NewCredentialRepo(deps.DB), // 凭证仓储
		Stats:       repository.NewStatsRepo(deps.DB),      // 统计仓储
		Compliance:  compRepo,                              // 合规仓储
		VCIssuer:    vc.NewIssuer(deps.Cfg.VCIssuerKey),    // 可验证凭证签发器
		OracleChain: deps.OracleChain,                      // 预言机链客户端
	}
	api.RegisterRoutes(r) // 注册所有 API 路由

	// 构建并返回服务器实例
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + deps.Port,  // 监听地址（所有接口 + 指定端口）
			Handler:      r,                // 路由处理器
			ReadTimeout:  15 * time.Second, // 读取超时设置为 15 秒
			WriteTimeout: 0,                // 写入超时设为 0（SSE 服务端推送需要长连接）
		},
	}
}

// ListenAndServe 启动 HTTP 服务器并监听连接
// 返回: 服务器关闭时的错误信息
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe() // 开始监听并处理请求
}

// Shutdown 优雅关闭 HTTP 服务器
// 等待现有连接处理完毕后再关闭
// 参数 ctx: 上下文，控制关闭超时
// 返回: 关闭过程中的错误信息
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx) // 执行优雅关闭
}

// Addr 返回服务器监听地址
// 返回: 监听地址字符串（如 ":8080"）
func (s *Server) Addr() string {
	return s.httpServer.Addr // 返回服务器地址
}

// String 返回服务器的可读 URL 字符串表示
// 返回: 完整的 URL 地址（如 "http://localhost:8080"）
func (s *Server) String() string {
	return fmt.Sprintf("http://localhost%s", s.Addr()) // 格式化为完整 URL
}
