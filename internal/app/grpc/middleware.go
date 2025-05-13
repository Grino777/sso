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
	"github.com/Grino777/sso/internal/services/auth"
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

type ReqMetadata struct {
	appID     int
	timestamp string
	secret    string
}

func HMACInterceptor(
	log *slog.Logger,
	db auth.DBStorage,
	cache auth.CacheStorage,
	mode string,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		switch mode {
		case "local":
			return handler(ctx, req)
		default:
			return validateHMAC(ctx, log, req, db, cache, handler)
		}
	}
}

// Валидирует запрос от приложений
func validateHMAC(ctx context.Context,
	sLog *slog.Logger,
	req any,
	db auth.DBStorage,
	cache auth.CacheStorage,
	handler grpc.UnaryHandler,
) (any, error) {
	const op = "grpcapp.middleware.validateHMAC"

	log := sLog.With("op", op)

	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		log.Error("metadata is empty")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	secret, err := appSecret(md)
	if err != nil {
		log.Error("error: %v", logger.Error(err))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	rm, err := extractMetadata(req)
	if err != nil {
		log.Error("error: %v", logger.Error(err))
		return nil, err
	}

	rm.secret = secret

	valid, err := validateSecret(rm, db, cache)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("app not found", "appID", rm.appID)
		}

		log.Error("failed to parse timestamp", "appID", rm.appID, "timestamp", rm.timestamp, "error", err.Error())

		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}
	if !valid {
		log.Error("invalid HMAC", "appID", rm.appID)

		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}

	log.Debug("HMAC validated",
		slog.Uint64("appID", uint64(rm.appID)),
		slog.String("timestamp", rm.timestamp),
	)

	return handler(ctx, req)
}

// Получает secret из запроса
func appSecret(md metadata.MD) (secret string, err error) {
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
		appID:     int(metadata.AppId),
		timestamp: metadata.Timestamp,
		secret:    "",
	}, nil

}

// Валидирует secret из запроса
func validateSecret(
	rm *ReqMetadata,
	db auth.DBStorage,
	cache auth.CacheStorage,
) (bool, error) {
	ts, err := time.Parse(time.RFC3339, rm.timestamp)
	if err != nil {
		return false, fmt.Errorf("failed parse timestamp: %w, %v", err, rm.timestamp)
	}

	ctx := context.Background()
	data := ts.Format(time.RFC3339) + strconv.Itoa(rm.appID)

	now := time.Now().UTC()
	if ts.Before(now.Add(-2*time.Minute)) || ts.After(now.Add(5*time.Second)) {
		return false, fmt.Errorf("%w: current time %v", errInvalidTimestamp, now.Format(time.RFC3339))
	}

	app, err := auth.GetCachedApp(ctx, db, cache, uint32(rm.appID))
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
