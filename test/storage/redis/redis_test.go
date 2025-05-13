package redis

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	appR "github.com/Grino777/sso/internal/storage/redis"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisStorage(t *testing.T) (*appR.RedisStorage, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	cfg := config.RedisConfig{
		Addr:        mr.Addr(),
		Password:    "",
		DB:          1,
		MaxRetries:  5,
		DialTimeout: 5 * time.Second,
		Timeout:     5 * time.Second,
		TokenTTL:    24 * time.Hour,
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store := &appR.RedisStorage{
		Cfg:        cfg,
		MaxRetries: 3,
		RetryDelay: time.Second,
		Log:        log,
		Client: redis.NewClient(&redis.Options{
			Addr:     mr.Addr(),
			Password: "",
			DB:       0,
		}),
	}

	return store, mr
}

func TestRedisStorage_SaveApp(t *testing.T) {
	// Определяем тестовые случаи
	tests := []struct {
		name        string
		app         *models.App
		prepare     func(*miniredis.Miniredis)       // Подготовка miniredis
		setupCtx    func(*testing.T) context.Context // Настройка контекста
		wantErr     bool
		expectedErr string
		verify      func(t *testing.T, store *appR.RedisStorage, app *models.App)
	}{
		{
			name: "successful save",
			app: &models.App{
				ID:     1,
				Name:   "test-app",
				Secret: "test-secret",
			},
			prepare: func(mr *miniredis.Miniredis) {},
			setupCtx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			wantErr: false,
			verify: func(t *testing.T, store *appR.RedisStorage, app *models.App) {
				// Проверяем, что приложение сохранено и может быть получено
				cachedApp, err := store.GetApp(context.Background(), uint32(app.ID))
				require.NoError(t, err)
				assert.Equal(t, app, cachedApp)
			},
		},
		{
			name: "invalid app ID",
			app: &models.App{
				ID:     0,
				Name:   "test-app",
				Secret: "test-secret",
			},
			prepare: func(mr *miniredis.Miniredis) {},
			setupCtx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			wantErr:     true,
			expectedErr: "invalid appID",
			verify: func(t *testing.T, store *appR.RedisStorage, app *models.App) {
				// Проверяем, что данные не сохранены
				_, err := store.GetApp(context.Background(), uint32(app.ID))
				assert.ErrorIs(t, err, appR.ErrCacheNotFound)
			},
		},
		{
			name: "redis error",
			app: &models.App{
				ID:     2,
				Name:   "test-app",
				Secret: "test-secret",
			},
			prepare: func(mr *miniredis.Miniredis) {
				mr.Close() // Закрываем miniredis для симуляции ошибки
			},
			setupCtx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			wantErr:     true,
			expectedErr: "failed to save app",
			verify: func(t *testing.T, store *appR.RedisStorage, app *models.App) {
				// Проверяем, что данные не сохранены
				_, err := store.GetApp(context.Background(), uint32(app.ID))
				assert.ErrorIs(t, err, appR.ErrCacheNotFound)
			},
		},
		// // {
		// // 	name: "context timeout",
		// // 	app: &models.App{
		// // 		ID:     3,
		// // 		Name:   "test-app",
		// // 		Secret: "test-secret",
		// // 	},
		// // 	prepare: func(mr *miniredis.Miniredis) {
		// // 		// Удаляем mr.SetDelay, так как метод не существует
		// // 	},
		// // 	setupCtx: func(t *testing.T) context.Context {
		// // 		// Увеличиваем таймаут до 1 мс, чтобы попытаться поймать таймаут
		// // 		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		// // 		t.Cleanup(cancel)
		// // 		return ctx
		// // 	},
		// // 	wantErr:     true,
		// // 	expectedErr: "context deadline exceeded",
		// // 	verify: func(t *testing.T, store *appR.RedisStorage, app *models.App) {
		// // 		// Проверяем, что данные не сохранены
		// // 		_, err := store.GetApp(context.Background(), uint32(app.ID))
		// // 		assert.ErrorIs(t, err, appR.ErrCacheNotFound)
		// // 	},
		// },
	}

	// Запускаем тесты
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mr := setupRedisStorage(t)
			defer mr.Close()
			defer store.Close()

			// Подготавливаем miniredis
			if tt.prepare != nil {
				tt.prepare(mr)
			}

			// Настраиваем контекст
			ctx := tt.setupCtx(t)

			// Выполняем SaveApp
			err := store.SaveApp(ctx, tt.app)

			// Проверяем ошибку
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			// Проверяем результат
			if tt.verify != nil {
				tt.verify(t, store, tt.app)
			}
		})
	}
}
