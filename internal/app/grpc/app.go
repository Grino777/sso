package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sso/internal/config"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gRPC server application
type App struct {
	cfg        *config.Config
	log        *slog.Logger
	gRPCServer *grpc.Server
}

// Create gRPC server object
func New(cfg *config.Config, log *slog.Logger) *App {

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	recoverOptions := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("error", err))

			return status.Error(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoverOptions...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
	))

	return &App{
		cfg:        cfg,
		log:        log,
		gRPCServer: gRPCServer,
	}
}

// InterceptorLogger adapts slog logger to interceptor logger.
// !!! This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// Run gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run gRPC Server
func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	a.log.Info("grpc server is running", slog.String("addr", (l.Addr().String())))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

// GracefulStop gRPC Server
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stoping gRPC server", slog.Int("port", int(a.cfg.GRPC.Port)))

	a.gRPCServer.GracefulStop()
}
