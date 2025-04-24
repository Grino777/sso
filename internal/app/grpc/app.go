package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	grpcauth "github.com/Grino777/sso/internal/grpc/server"
	"github.com/Grino777/sso/internal/storage"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Объект приложения для управления GRPC сервром
type GRPCApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
	mode       string
}

func New(
	log *slog.Logger,
	authService grpcauth.AuthService,
	dbstore storage.Storage,
	redisstore storage.Storage,
	port int,
	mode string,
) *GRPCApp {

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
		HMACInterceptor(log, dbstore, mode),
	))

	grpcauth.RegServer(gRPCServer, authService)

	return &GRPCApp{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
		mode:       mode,
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
func (a *GRPCApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run gRPC Server
func (a *GRPCApp) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	a.log.Info("grpc server is running",
		slog.String("addr", (l.Addr().String())),
		slog.String("mode", a.mode),
	)

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

// GracefulStop gRPC Server
func (a *GRPCApp) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stoping gRPC server", slog.Int("port", int(a.port)))

	a.gRPCServer.GracefulStop()
}
