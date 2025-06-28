package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"time"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errAppSecret        = errors.New("secret key not found")
	errInvalidTimestamp = errors.New("invalid timestamp")
	errInvalidArgs      = status.Error(codes.Unauthenticated, "invalid or missing authorization")
)

// ReqMetadata содержит данные из заголовка запроса
type ReqMetadata struct {
	appID     uint64
	timestamp string
	secret    string
}

// InterceptorLogger adapts slog logger to interceptor logger.
// !!! This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// HMACInterceptor проверяет HMAC-токен в заголовке запроса
func HMACInterceptor(
	log *slog.Logger,
	services Services,
	mode string,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		const op = opGrpc + "HMACInterceptor"

		log := log.With(slog.String("op", op))

		switch mode {
		case "local":
			return handler(ctx, req)
		default:
			return validateHMAC(ctx, log, req, services, handler)
		}
	}
}

// validateHMAC проверяет HMAC-токен в заголовке запроса
func validateHMAC(ctx context.Context,
	log *slog.Logger,
	req any,
	services Services,
	handler grpc.UnaryHandler,
) (any, error) {
	const op = opGrpc + "validateHMAC"

	log = log.With("op", op)

	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		log.Error("metadata is empty")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	secret, err := getMdAppSecret(md)
	if err != nil {
		log.Error("failed to get app secret from metadata : %w", logger.Error(err))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	rm, err := extractMetadata(req)
	if err != nil {
		log.Error("failed to extract metadata", logger.Error(err))
		return nil, err
	}

	rm.secret = secret
	log = log.With(slog.Uint64("app_id", rm.appID))

	valid, err := validateSecret(rm, services)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("app not found", logger.Error(err))
		}

		log.Error(
			"failed to parse timestamp",
			slog.String("timestamp", rm.timestamp),
			logger.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}
	if !valid {
		log.Error("invalid HMAC")
		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}

	return handler(ctx, req)
}

// Получает secret из запроса
func getMdAppSecret(md metadata.MD) (secret string, err error) {
	appSecret, exist := md["authorization"]
	if !exist || len(appSecret) == 0 {
		return "", errAppSecret
	}

	return appSecret[0], nil
}

// Получает данные из metadata запроса
func extractMetadata(req any) (rm *ReqMetadata, err error) {

	v := reflect.ValueOf(req)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil, errInvalidArgs
	}

	metadataField := v.Elem().FieldByName("Metadata")
	if !metadataField.IsValid() || metadataField.IsNil() {
		return nil, errInvalidArgs
	}

	metadata := metadataField.Interface().(*sso_v1.AuthMetadata)

	return &ReqMetadata{
		appID:     uint64(metadata.AppId),
		timestamp: metadata.Timestamp,
		secret:    "",
	}, nil

}

// Валидирует secret из запроса
func validateSecret(
	rm *ReqMetadata,
	services Services,
) (bool, error) {
	const op = opGrpc + "validateSecret"

	ts, err := time.Parse(time.RFC3339, rm.timestamp)
	if err != nil {
		return false, fmt.Errorf("%s: failed parse timestamp %s: %w", op, rm.timestamp, err)
	}

	ctx := context.Background()
	data := ts.Format(time.RFC3339) + strconv.FormatUint(rm.appID, 10)

	now := time.Now().UTC()
	if ts.Before(now.Add(-2*time.Minute)) || ts.After(now.Add(5*time.Second)) {
		return false, fmt.Errorf("%w: current time %s", errInvalidTimestamp, now.Format(time.RFC3339))
	}

	auth := services.Auth()

	app, err := auth.GetCachedApp(ctx, uint32(rm.appID))
	if err != nil {
		return false, err
	}

	expectedHMAC := computeHMAC(data, app.Secret)

	return expectedHMAC == rm.secret, nil
}

// Вычисляет HMAC
func computeHMAC(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
