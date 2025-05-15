package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Константы для режимов окружения
const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// Константы для путей к конфигурационным файлам
const (
	localConfigPath = "config/local.yaml"
	devConfigPath   = "config/dev.yaml"
	prodConfigPath  = "config/prod.yaml"
)

// Константы для переменных окружения
const (
	dbUserEnv  = "DB_USER"
	dbPassEnv  = "DB_PASSWORD"
	keysDirEnv = "KEYS_DIR"
)

// Config представляет конфигурацию приложения.
type Config struct {
	Mode            string        // Режим работы приложения (local, dev, prod)
	DB              DBConfig      `yaml:"db" env-required:"true"`
	Redis           RedisConfig   `yaml:"redis" env-required:"true"`
	GRPC            GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTTL        time.Duration `yaml:"tokenTTL" env-default:"1h"`
	RefreshTokenTTL time.Duration `yaml:"refreshTokenTTL" env-default:"168h"`
	DBUser          DBUser
	BaseDir         string
	KeysDir         string
}

// DBUser содержит учетные данные для подключения к базе данных.
type DBUser struct {
	User     string
	Password string
}

// DBConfig содержит настройки базы данных.
type DBConfig struct {
	StoragePath string `yaml:"storage_path" env-required:"true"`
}

// GRPCConfig содержит настройки gRPC-сервера.
type GRPCConfig struct {
	URL     string        `yaml:"url" env-required:"true"`
	Port    uint16        `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

// RedisConfig содержит настройки Redis.
type RedisConfig struct {
	Addr        string        `yaml:"addr" env-default:"127.0.0.1:6379"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db" env-default:"0"`
	MaxRetries  int           `yaml:"max_retries" env-default:"5"`
	DialTimeout time.Duration `yaml:"dial_timeout" env-default:"10s"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	TokenTTL    time.Duration `yaml:"tokenTTL" env-default:"1h"`
}

var envMapping = map[string]func(*Config, string){
	dbUserEnv:  func(c *Config, v string) { c.DBUser.User = v },
	dbPassEnv:  func(c *Config, v string) { c.DBUser.Password = v },
	keysDirEnv: func(c *Config, v string) { c.KeysDir = v },
}

var (
	// Глобальные переменные для кэширования конфигурации
	config    *Config
	configErr error
	once      sync.Once
	flagSet   *flag.FlagSet
	mode      string
)

func init() {
	flagSet = flag.NewFlagSet("sso", flag.ContinueOnError)
	flagSet.StringVar(&mode, "mode", "", "application mode (local, dev, prod)")
}

// Load загружает конфигурацию приложения из флагов, переменных окружения и YAML-файла.
// Вызывается однократно благодаря sync.Once.
func Load() (*Config, error) {
	const op = "config.Load"

	once.Do(func() {
		config, configErr = loadConfig()
	})

	if configErr != nil {
		return nil, fmt.Errorf("%s: %w", op, configErr)
	}

	return config, nil
}

func loadConfig() (*Config, error) {
	const op = "config.loadConfig"

	cfg := &Config{}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("%s: failed to parse flags: %w", op, err)
	}

	if mode == "" {
		return nil, fmt.Errorf("%s: mode not specified", op)
	}
	cfg.Mode = mode

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: failed to load .env file: %w", op, err)
	}

	configPath, err := configPath(cfg.Mode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: config file does not exist: %s", op, configPath)
		}
		return nil, fmt.Errorf("%s: cannot stat config file: %w", op, err)
	}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to read config file: %w", op, err)
	}

	if err := parseEnv(cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse environment variables: %w", op, err)
	}

	if err := setBaseDir(cfg, configPath); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	setKeysDir(cfg)

	return cfg, nil
}

func configPath(mode string) (string, error) {
	const op = "config.configPath"

	switch mode {
	case EnvLocal:
		return localConfigPath, nil
	case EnvDev:
		return devConfigPath, nil
	case EnvProd:
		return prodConfigPath, nil
	default:
		return "", fmt.Errorf("%s: invalid mode: %s", op, mode)
	}
}

func parseEnv(cfg *Config) error {
	const op = "config.parseEnv"

	for envKey, setter := range envMapping {
		value, err := getEnvValue(envKey)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		setter(cfg, value)
	}
	return nil
}

func getEnvValue(key string) (string, error) {
	const op = "config.getEnvValue"

	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s: environment variable %s not set", op, key)
	}
	return value, nil
}

func setBaseDir(cfg *Config, configPath string) error {
	const op = "config.setBaseDir"

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("%s: cannot get absolute path: %w", op, err)
	}

	cfg.BaseDir = filepath.Dir(filepath.Dir(absPath))
	return nil
}

func setKeysDir(cfg *Config) {
	cfg.KeysDir = filepath.Join(cfg.BaseDir, "keys")
}
