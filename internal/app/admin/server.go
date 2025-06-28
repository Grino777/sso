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

const opAdmin = "app.admin."

type Certificate struct {
	certPath string
	keyPath  string
}

type APIServer struct {
	Logger *slog.Logger
	Router *gin.Engine
	Routes *Routes
	Server *http.Server
	Config config.ApiServerConfig
}

// NewApiServer создает новый экземпляр APIServer
func NewApiServer(log *slog.Logger, cfg config.ApiServerConfig, keysStore keysStore) *APIServer {
	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
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
	engine.Use(gin.Recovery())

	server := &http.Server{
		Addr:    cfg.Addr + ":" + cfg.Port,
		Handler: engine,
	}

	routes := NewRoutes(keysStore)
	routes.RegisterRoutes(engine)

	return &APIServer{
		Logger: log,
		Router: engine,
		Server: server,
		Config: cfg,
	}
}

// Run запускает сервер
func (as *APIServer) Run(ctx context.Context) error {
	const op = opAdmin + "Run"

	errChan := make(chan error, 1)

	cert, err := as.init()
	if err != nil {
		return err
	}

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

// Stop останавливает сервер
func (as *APIServer) Stop() error {
	const op = opAdmin + "Stop"

	log := as.Logger.With(slog.String("op", op))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := as.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: failed to shutdown api server: %w", op, err)
	}

	log.Debug("api server successfully stopped")

	return nil
}

// init Проверяет существование директории и файлов сертификата
// и создает сертификаты если их нет
func (as *APIServer) init() (Certificate, error) {
	const op = opAdmin + "init"

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
