package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/utils/certs"
	"github.com/gin-gonic/gin"
)

const opServer = "app.admin."

type Certificate struct {
	certPath string
	keyPath  string
}

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
				slog.String("client_ip", param.ClientIP),
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
	const op = opServer + "RegisterRoutes"

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
	const op = opServer + "Run"

	errChan := make(chan error, 1)

	cert, err := as.init()
	if err != nil {
		return err
	}

	as.RegisterRoutes()

	go func() {
		if err := as.Server.ListenAndServeTLS(cert.certPath, cert.keyPath); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("%s: failed to starting api server: %w", op, err)
		}
	}()

	var errs []error

	select {
	case <-ctx.Done():
	case err := <-errChan:
		errs = append(errs, err)
	}

	if err := as.Stop(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (as *APIServer) Stop() error {
	const op = opServer + "Stop"

	log := as.Logger.With(slog.String("op", op))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := as.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: failed to shutdown api server: %w", op, err)
	}

	log.Debug("api server successfully stopped")

	return nil
}

func (as *APIServer) init() (Certificate, error) {
	const op = opServer + "init"

	var certObj Certificate

	if err := certs.CheckCertsFolder(as.Config.CertsDir); err != nil {
		return certObj, fmt.Errorf("%s: failed to checking certificate folder: %w", op, err)
	}

	expired, err := certs.CheckCertificate(as.Config.CertsDir)
	if err != nil {
		return certObj, fmt.Errorf("%s: failed to checking certificate: %w", op, err)
	}
	if expired {
		if err := certs.CreateCertsFiles(as.Config.CertsDir); err != nil {
			return certObj, fmt.Errorf("%s: failed to creating certificate: %w", op, err)
		}
	}

	certPath := filepath.Join(as.Config.CertsDir, "cert.pem")
	keyPath := filepath.Join(as.Config.CertsDir, "key.pem")

	certObj.certPath = certPath
	certObj.keyPath = keyPath

	return certObj, nil
}
