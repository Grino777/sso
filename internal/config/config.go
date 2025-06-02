package config

import (
	"errors"
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
	LocalMode = "local"
	DevMode   = "dev"
	ProdMode  = "prod"
)

// Константы для путей к конфигурационным файлам
const (
	localConfigPath = "config/local.yaml"
	devConfigPath   = "config/dev.yaml"
	prodConfigPath  = "config/prod.yaml"
)

// Константы для переменных окружения
const (
	suUsernameEnv = "DB_USER"
	suPassEnv     = "DB_PASSWORD"
	keysDirEnv    = "KEYS_DIR"
)

// Константы с кредами для Postgres
const (
	pgUser = "PG_USER"
	pgPass = "PG_PASS"
	pgHost = "PG_HOST"
	pgPort = "PG_PORT"
)

// Константы для типов баз данных
const (
	DBTypePostgres = "postgres"
	DBTypeSQLite   = "sqlite"
)

const configOp = "config.config."

var (
	ErrModeFlag = errors.New("invalid mode flag")
	ErrDbFlag   = errors.New("invalid db flag")
)

var (
	// Глобальные переменные для кэширования конфигурации
	config    *Config
	configErr error
	once      sync.Once
	flagSet   *flag.FlagSet
	modeFlag  string
	dbFlag    string
)

var envMapping = map[string]func(*Config, string){
	suUsernameEnv: func(c *Config, v string) { c.SuperUser.Username = v },
	suPassEnv:     func(c *Config, v string) { c.SuperUser.Password = v },
	keysDirEnv:    func(c *Config, v string) { c.FS.KeysDir = v },
}

var envPGMapping = map[string]func(*Config, string){
	pgUser: func(c *Config, v string) { c.Database.DBUser = v },
	pgPass: func(c *Config, v string) { c.Database.DBPass = v },
	pgHost: func(c *Config, v string) { c.Database.DBHost = v },
	pgPort: func(c *Config, v string) { c.Database.DBPort = v },
}

// Config представляет конфигурацию приложения.
type Config struct {
	Mode      string
	GRPC      GRPCConfig `yaml:"grpc" env-required:"true"`
	Database  DatabaseConfig
	Redis     RedisConfig `yaml:"redis" env-required:"true"`
	Tokens    TokenConfig
	FS        FileSystemConfig
	SuperUser SuperUser
	ApiServer ApiServerConfig `yaml:"api_server" env-required:"true"`
}

type DatabaseConfig struct {
	DBType           string
	DBName           string `yaml:"db_name" env-default:"sso"`
	LocalStoragePath string `yaml:"local_storage_path" env-default:"./storage/sso.sqlite3"`
	DBUser           string
	DBPass           string
	DBHost           string
	DBPort           string
}

type TokenConfig struct {
	TokenTTL        time.Duration `yaml:"tokenTTL" env-default:"1h"`
	RefreshTokenTTL time.Duration `yaml:"refreshTokenTTL" env-default:"168h"`
}

type FileSystemConfig struct {
	BaseDir    string
	KeysDir    string
	ConfigPath string
}

// Пользователь с ролью superuser для взаимодействия с GRPC
type SuperUser struct {
	Username string
	Password string
}

// GRPCConfig содержит настройки gRPC-сервера.
type GRPCConfig struct {
	Addr    string        `yaml:"grpc_addr" env-required:"true"`
	Port    uint16        `yaml:"grpc_port" env-required:"true"`
	Timeout time.Duration `yaml:"grpc_timeout" env-default:"5s"`
}

// RedisConfig содержит настройки Redis.
type RedisConfig struct {
	Addr        string        `yaml:"redis_addr" env-default:"127.0.0.1:6379"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db" env-default:"0"`
	MaxRetries  int           `yaml:"max_retries" env-default:"5"`
	DialTimeout time.Duration `yaml:"dial_timeout" env-default:"10s"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	TokenTTL    time.Duration `yaml:"tokenTTL" env-default:"1h"`
}

type ApiServerConfig struct {
	Addr     string `yaml:"api_addr" env-required:"true"`
	Port     string `yaml:"api_port" env-required:"true"`
	CertsDir string
}

// GetFlagSet возвращает flagSet для использования в других пакетах.
func GetFlagSet() *flag.FlagSet {
	return flagSet
}

// GetModeFlag возвращает значение флага mode.
func GetModeFlag() string {
	return modeFlag
}

// GetDbFlag возвращает значение флага db.
func GetDbFlag() string {
	return dbFlag
}

// Load загружает конфигурацию приложения из флагов, переменных окружения и YAML-файла.
// Вызывается однократно благодаря sync.Once.
func Load() (*Config, error) {
	const op = configOp + "Load"

	once.Do(func() {
		config, configErr = loadConfig()
	})

	if configErr != nil {
		return nil, fmt.Errorf("%s: %w", op, configErr)
	}

	return config, nil
}

func parseFlag(cfg *Config) error {
	const op = configOp + "parseFlag"

	flagSet = flag.NewFlagSet("sso", flag.ContinueOnError)
	flagSet.StringVar(&modeFlag, "mode", "", "application mode (local, dev, prod)")
	flagSet.StringVar(&dbFlag, "db", "", "database type (postgres, sqlite)")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("%s: failed to parse flags: %w", op, err)
	}
	if err := validateModeFlag(modeFlag, cfg); err != nil {
		return err
	}
	if err := validateDbFlag(dbFlag, cfg); err != nil {
		return err
	}
	return nil
}

func validateModeFlag(modeFlag string, cfg *Config) error {
	// const op = configOp + "validateMode"

	switch modeFlag {
	case LocalMode, DevMode, ProdMode:
		cfg.Mode = modeFlag
	default:
		return ErrModeFlag
	}
	return nil
}

func validateDbFlag(dbFlag string, cfg *Config) error {
	// const op = configOp + "validateDbFlag"

	switch dbFlag {
	case DBTypePostgres, DBTypeSQLite:
		cfg.Database.DBType = dbFlag
	default:
		return ErrDbFlag
	}
	return nil
}

func loadConfig() (*Config, error) {
	const op = configOp + "loadConfig"

	cfg := &Config{}

	if err := parseFlag(cfg); err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: failed to load .env file: %w", op, err)
	}

	if err := configPath(cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := os.Stat(cfg.FS.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: config file does not exist: %s", op, cfg.FS.ConfigPath)
		}
		return nil, fmt.Errorf("%s: cannot stat config file: %w", op, err)
	}

	if err := cleanenv.ReadConfig(cfg.FS.ConfigPath, cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to read config file %q: %w", op, cfg.FS.ConfigPath, err)
	}

	if err := parseEnv(cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse environment variables: %w", op, err)
	}

	if err := setBaseDir(cfg, cfg.FS.ConfigPath); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	cfg.FS.KeysDir = filepath.Join(cfg.FS.BaseDir, "keys")
	cfg.ApiServer.CertsDir = filepath.Join(cfg.FS.BaseDir, "certs")

	return cfg, nil
}

func configPath(cfg *Config) error {
	const op = configOp + "configPath"

	switch cfg.Mode {
	case LocalMode:
		cfg.FS.ConfigPath = localConfigPath
	case DevMode:
		cfg.FS.ConfigPath = devConfigPath
	case ProdMode:
		cfg.FS.ConfigPath = prodConfigPath
	default:
		return fmt.Errorf("%s: invalid mode: %s", op, cfg.Mode)
	}
	return nil
}

func parseEnv(cfg *Config) error {
	const op = configOp + "parseEnv"

	for envKey, setter := range envMapping {
		value, err := getEnvValue(envKey)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		setter(cfg, value)
	}

	if cfg.Database.DBType == DBTypePostgres {
		if err := parseEnvPG(cfg); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func parseEnvPG(cfg *Config) error {
	const op = configOp + "parsePgVariables"

	for envKey, setter := range envPGMapping {
		value, err := getEnvValue(envKey)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		setter(cfg, value)
	}
	return nil
}

func getEnvValue(key string) (string, error) {
	const op = configOp + "getEnvValue"

	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s: environment variable %s not set", op, key)
	}
	return value, nil
}

func setBaseDir(cfg *Config, configPath string) error {
	const op = configOp + "setBaseDir"

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("%s: cannot get absolute path: %w", op, err)
	}

	cfg.FS.BaseDir = filepath.Dir(filepath.Dir(absPath))
	return nil
}
