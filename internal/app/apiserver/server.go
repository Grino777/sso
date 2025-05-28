package apiserver

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/utils/certs"
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	Logger *slog.Logger
	Router *gin.Engine
	Server *http.Server
	Config config.ApiServerConfig
}

func NewApiServer(log *slog.Logger, cfg config.ApiServerConfig) *APIServer {
	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			log.Info("Gin request",
				slog.String("method", param.Method),
				slog.String("path", param.Path),
				slog.Int("status", param.StatusCode),
				slog.String("latency", param.Latency.String()),
			)
			return ""
		},
	}))
	router.Use(gin.Recovery())

	server := &http.Server{
		Addr:    cfg.Addr + ":" + cfg.Port,
		Handler: router,
	}

	return &APIServer{
		Logger: log,
		Router: router,
		Server: server,
		Config: cfg,
	}
}

func (as *APIServer) RegisterRoutes() {
	const op = "app.apiserver.RegisterRoutes"

	log := as.Logger.With(slog.String("op", op))

	for _, route := range routes {
		switch route.method {
		case "GET":
			as.Router.GET(route.path, route.handler)
		case "POST":
			as.Router.POST(route.path, route.handler)
		}
		log.Debug("route successfully registred",
			slog.String("method", route.method),
			slog.String("path", route.path))
	}
}

func (as *APIServer) Run(ctx context.Context) error {
	const op = "app.apiserver.Run"

	log := as.Logger.With("op", op)

	expired, err := certs.CheckCertificate(as.Config.CertsDir)
	if err != nil {
		log.Error("failed to checking certificate: %v", logger.Error(err))
		return err
	}
	if expired {
		if err := certs.CreateCertsFiles(as.Config.CertsDir); err != nil {
			log.Error("failed to creating certificate: %v", logger.Error(err))
			return err
		}
	}

	certPath := filepath.Join(as.Config.CertsDir, "cert.pem")
	keyPath := filepath.Join(as.Config.CertsDir, "key.pem")

	if err := as.Server.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
		log.Error("failed to starting api server", logger.Error(err))
		return err
	}

	<-ctx.Done()
	as.Logger.Info("shutting down api server")

	as.Stop()
	as.Logger.Info("api server stopped")

	return nil
}

func (as *APIServer) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := as.Server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
