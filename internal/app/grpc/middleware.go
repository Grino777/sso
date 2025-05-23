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

	"github.com/Grino777/sso/internal/storage"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
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

func HMACInterceptor(log *slog.Logger, dbStore storage.Storage, mode string) grpc.UnaryServerInterceptor {
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
			return validateHMAC(ctx, log, req, dbStore, handler)
		}
	}
}

func validateHMAC(ctx context.Context,
	log *slog.Logger,
	req any,
	dbStore storage.Storage,
	handler grpc.UnaryHandler,
) (any, error) {
	const op = "grpcapp.middleware.processReq"
	md, exist := metadata.FromIncomingContext(ctx)
	if !exist {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	secret, err := appSecret(md)
	if err != nil {
		log.Error("%s: %v", op, err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	appID, timestamp, err := requestMetadata(req)
	if err != nil {
		log.Error("%s: %v", op, err)
		return nil, err
	}

	valid, err := validateSecret(appID, timestamp, secret, dbStore)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("app not found", "appID", appID, "error", err.Error())
		}

		log.Error("failed to parse timestamp", "appID", appID, "timestamp", timestamp, "error", err.Error())

		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}
	if !valid {
		log.Error("invalid HMAC", "appID", appID)

		return nil, status.Error(codes.Unauthenticated, "invalid data transmitted")
	}

	log.Debug("HMAC validated",
		slog.Uint64("appID", uint64(appID)),
		slog.String("timestamp", timestamp),
	)

	return handler(ctx, req)
}

func appSecret(md metadata.MD) (secret string, err error) {
	appSecret, exist := md["authorization"]
	if !exist || len(appSecret) == 0 {
		return "", errAppSecret
	}

	return appSecret[0], nil
}

func requestMetadata(req any) (appID uint, timestamp string, err error) {

	v := reflect.ValueOf(req)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return 0, "", errInvalidArgs
	}

	metadataField := v.Elem().FieldByName("Metadata")
	if !metadataField.IsValid() || metadataField.IsNil() {
		return 0, "", errInvalidArgs
	}

	metadata := metadataField.Interface().(*sso_v1.AuthMetadata)

	return uint(metadata.AppId), metadata.Timestamp, nil
}

func computeHMAC(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func validateSecret(
	appID uint,
	timestamp string,
	secret string,
	dbStore storage.Storage,
) (bool, error) {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false, fmt.Errorf("failed parse timestamp: %w, %v", err, timestamp)
	}

	ctx := context.Background()
	data := ts.Format(time.RFC3339) + strconv.Itoa(int(appID))

	now := time.Now().UTC()
	if ts.Before(now.Add(-2*time.Minute)) || ts.After(now.Add(5*time.Second)) {
		return false, fmt.Errorf("%w: current time %v", errInvalidTimestamp, now.Format(time.RFC3339))
	}

	app, err := dbStore.Apps.GetApp(ctx, uint32(appID))
	if err != nil {
		return false, err
	}

	expectedHMAC := computeHMAC(data, app.Secret)

	return expectedHMAC == secret, nil
}
