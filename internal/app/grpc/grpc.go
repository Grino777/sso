package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/Grino777/sso/internal/config"
	grpcauth "github.com/Grino777/sso/internal/delivery/grpc/auth"
	grpcjwks "github.com/Grino777/sso/internal/delivery/grpc/jwks"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"

	_ "net/http/pprof"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const opGrpc = "app.grpc."

// Сервисы приложения
type Services interface {
	Auth() *auth.AuthService
	Jwks() *jwks.JwksService
}

// Объект приложения для управления GRPC сервром
type GRPCApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
	mode       string
}

// NewGrpcApp создает новый объект приложения для управления GRPC сервером
func NewGrpcApp(
	log *slog.Logger,
	services Services,
	cfg *config.Config,
) *GRPCApp {

	logger := log.With(slog.String("port", fmt.Sprint(cfg.GRPC.Port)))

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	recoverOptions := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("Recovered from panic", slog.Any("error", err))

			return status.Error(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoverOptions...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
		HMACInterceptor(log, services, cfg.Mode),
	))

	grpcauth.RegServer(gRPCServer, services.Auth())
	grpcjwks.RegService(gRPCServer, services.Jwks())

	return &GRPCApp{
		log:        logger,
		gRPCServer: gRPCServer,
		port:       int(cfg.GRPC.Port),
		mode:       cfg.Mode,
	}
}

// Run запускает GRPC-сервер
func (a *GRPCApp) Run(ctx context.Context) error {
	const op = opGrpc + "Run"

	var errs []error

	log := a.log.With(slog.String("op", op))
	errChan := make(chan error, 1)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("server failed to start listening on port %d: %w", a.port, err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.Error("failed to close listener", logger.Error(err))
			errs = append(errs, err)
		}
		log.Debug("listener successfully closed")
	}()

	go func() {
		if err := a.gRPCServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Error("failed to serve grpc server", logger.Error(err))
			errChan <- fmt.Errorf("%s: %w", op, err)
		}
	}()

	log.Info("grpc server is running")

	select {
	case <-ctx.Done(): // Context canceled
	case err := <-errChan:
		errs = append(errs, err)
	}

	a.Stop()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// GracefulStop for gRPC Server
func (a *GRPCApp) Stop() {
	const op = "grpcapp.Stop"

	a.gRPCServer.GracefulStop()
	a.log.Info("gRPC server stopped", slog.String("op", op))
}
