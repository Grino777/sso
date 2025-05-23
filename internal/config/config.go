package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	EnvLocal    = "local"
	EnvDev      = "dev"
	EnvProd     = "prod"
	localConfig = "config/local.yaml"
	devConfig   = "config/dev.yaml"
	prodConfig  = "config/prod.yaml"
	db_user     = "DB_USER"
	db_pass     = "DB_PASSWORD"
)

type Config struct {
	Mode     string
	DB       DBConfig      `yaml:"db" env-required:"true"`
	Redis    RedisConfig   `yaml:"redis" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTTL time.Duration `yaml:"tokenTTL" env-default:"1h"`
	DBUser   DBUser
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
	Addr        string        `yaml:"addr" default:"127.0.0.1"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db" default:"0"`
	MaxRetries  int           `yaml:"max_retries" default:"5"`
	DialTimeout time.Duration `yaml:"dial_timeout" default:"10"`
	Timeout     time.Duration `yaml:"timeout" default:"5"`
}

// Загрузка конфигураций и .env для приложения
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		panic(".env file not found or failed to load")
	}

	cfgPath := path()
	cfg := mustParseConfig(cfgPath)
	return cfg
}

// Возвращает путь до файла с необходимыми конфигами
func path() string {
	m := parseMode()
	configPath := configPath(m)

	return configPath
}

// Парсит конфиг файл, при отсутсвии файла вызывает panic
func mustParseConfig(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config:" + err.Error())
	}

	dbUser := parseEnv()
	cfg.DBUser = dbUser

	return &cfg

}

func parseEnv() DBUser {
	user := parseEnvValue(db_user)
	pass := parseEnvValue(db_pass)

	return DBUser{
		User:     user,
		Password: pass,
	}
}

func parseEnvValue(key string) string {
	res := os.Getenv(key)
	if res != "" {
		return res
	}
	errString := fmt.Sprintf("%s not specifed in env", res)
	panic(errString)
}

// Парсит mode из переданных args во время запуска приложения
func parseMode() string {
	var res string

	flag.StringVar(&res, "mode", "", "app mode")
	flag.Parse()

	if res == "" {
		panic("'mode' was not specified")
	}
	return res
}

// Выбирает путь для конфига из заданных констант и паникует при неизвестном режиме
func configPath(mode string) string {
	switch mode {
	case EnvLocal:
		return localConfig
	case EnvDev:
		return devConfig
	case EnvProd:
		return prodConfig
	default:
		panic("mode is incorrect")
	}
}
