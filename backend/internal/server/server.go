package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/handler"
	appmw "github.com/prediction-did/simple/internal/middleware"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/vc"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	httpServer *http.Server
}

type Dependencies struct {
	Port        string
	Cfg         *config.Config
	DB          *pgxpool.Pool
	Redis       *redis.Client
	Chain       *blockchain.Client
	OracleChain *blockchain.OracleClient
}

func New(deps Dependencies) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmw.RateLimit(deps.Cfg.RateLimitPerMinute))

	compRepo := repository.NewComplianceRepo(deps.DB)
	r.Use(appmw.GeoBlock(deps.Cfg, func(ip, country, path string, allowed bool) {
		_ = compRepo.LogGeo(context.Background(), ip, country, path, allowed)
	}))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Admin-Key", "CF-IPCountry", "X-Country-Code", "X-KYC-Signature"},
		AllowCredentials: true,
	}))

	health := &handler.Health{
		DB:    deps.DB,
		Redis: deps.Redis,
		Chain: deps.Chain,
	}
	health.RegisterRoutes(r)

	api := &handler.API{
		Cfg:         deps.Cfg,
		Matches:     repository.NewMatchRepo(deps.DB),
		Markets:     repository.NewMarketRepo(deps.DB),
		Users:       repository.NewUserRepo(deps.DB),
		Positions:   repository.NewPositionRepo(deps.DB),
		OracleJobs:  repository.NewOracleJobRepo(deps.DB),
		Credentials: repository.NewCredentialRepo(deps.DB),
		Stats:       repository.NewStatsRepo(deps.DB),
		Compliance:  compRepo,
		VCIssuer:    vc.NewIssuer(deps.Cfg.VCIssuerKey),
		OracleChain: deps.OracleChain,
	}
	api.RegisterRoutes(r)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + deps.Port,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 0, // SSE needs long write
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}

func (s *Server) String() string {
	return fmt.Sprintf("http://localhost%s", s.Addr())
}
