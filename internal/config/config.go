package config

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	// Environment modes
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"

	// Config file paths
	localConfig = "config/local.yaml"
	devConfig   = "config/dev.yaml"
	prodConfig  = "config/prod.yaml"

	// Environment variables
	db_user = "DB_USER"
	db_pass = "DB_PASSWORD"
	keysDir = "KEYS_DIR"
)

var (
	envVariables = []string{db_user, db_pass, keysDir}
	envMappings  = map[string]func(*Config, string){
		db_user: func(c *Config, v string) { c.DBUser.User = v },
		db_pass: func(c *Config, v string) { c.DBUser.Password = v },
		keysDir: func(c *Config, v string) { c.KeysDir = v },
	}
)

type Config struct {
	Mode            string
	DB              DBConfig      `yaml:"db" env-required:"true"`
	Redis           RedisConfig   `yaml:"redis" env-required:"true"`
	GRPC            GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTTL        time.Duration `yaml:"tokenTTL" env-default:"1h"`
	RefreshTokenTTL time.Duration `yaml:"refreshTokenTTL" env-default:"7d"`
	DBUser          DBUser
	BaseDir         string
	KeysDir         string
}

type DBUser struct {
	User     string
	Password string
}

type DBConfig struct {
	Storage_path string `yaml:"storage_path" env-required:"true"`
}

type GRPCConfig struct {
	Url     string        `yaml:"url" env-required:"true"`
	Port    uint16        `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type RedisConfig struct {
	Addr        string        `yaml:"addr" default:"127.0.0.1:6379"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db" default:"0"`
	MaxRetries  int           `yaml:"max_retries" default:"5"`
	DialTimeout time.Duration `yaml:"dial_timeout" default:"10"`
	Timeout     time.Duration `yaml:"timeout" default:"5"`
	TokenTTL    time.Duration `yaml:"tokenTTL" default:"1h"`
}

// Загрузка конфигураций и .env для App
func Load() (*Config, error) {
	const op = "config.Load"

	cfg := &Config{}

	cfg, err := mustParseConfig()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to load config: %v", op, err)
	}

	return cfg, nil
}

func mustParseConfig() (*Config, error) {
	const op = "config.mustParseConfig"

	cfg := &Config{}

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	cfgPath, err := resolveConfigPath()
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	_, err = os.Stat(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: config file does not exist '%s': %v", op, cfgPath, err)
		}
		return nil, fmt.Errorf("%s: cannot stat config file: %v", op, err)
	}

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return nil, fmt.Errorf("%s: cannot read config: %v", op, err)
	}

	if err := parseEnv(cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse environment variables: %v", op, err)
	}

	if err := getBaseDir(cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to get base directory: %v", op, err)
	}

	getKeysDir(cfg)

	return cfg, nil

}

func resolveConfigPath() (string, error) {
	const op = "config.getPath"

	m, err := parseMode()
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}

	configPath, err := configPath(m)
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}

	return configPath, nil
}

func parseEnv(cfg *Config) error {
	const op = "config.parseEnv"

	for _, envKey := range envVariables {
		value, err := getEnvValue(envKey)
		if err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
		setter, ok := envMappings[envKey]
		if !ok {
			return fmt.Errorf("%s: no mapping for environment variable: %s", op, envKey)
		}
		setter(cfg, value)
	}
	return nil
}

func getEnvValue(key string) (string, error) {
	const op = "config.parseEnvValue"

	res := os.Getenv(key)
	if res != "" {
		return res, nil
	}

	return "", fmt.Errorf("%s: not specifed in env `%s`", op, res)
}

func parseMode() (string, error) {
	const op = "config.parseMode"

	var res string

	flag.StringVar(&res, "mode", "", "app mode")
	flag.Parse()

	if res == "" {
		return "", fmt.Errorf("%s: 'mode' was not specified", op)
	}

	return res, nil
}

func configPath(mode string) (string, error) {
	const op = "config.configPath"

	switch mode {
	case EnvLocal:
		return localConfig, nil
	case EnvDev:
		return devConfig, nil
	case EnvProd:
		return prodConfig, nil
	default:
		return "", fmt.Errorf("%s: incorrect mode: %s", op, mode)
	}
}

func getBaseDir(c *Config) error {
	const op = "config.getBaseDir"

	configPath, err := resolveConfigPath()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("%s: cannot get absolute path: %v", op, err)
	}

	baseDir := filepath.Dir(filepath.Dir(absPath))
	c.BaseDir = baseDir

	return nil
}

func getKeysDir(c *Config) {
	path := path.Join(c.BaseDir, "keys")
	c.KeysDir = path
}
