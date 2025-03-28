package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal    = "local"
	EnvDev      = "dev"
	EnvProd     = "prod"
	localConfig = "config/local.yaml"
	devConfig   = "config/dev.yaml"
	prodConfig  = "config/prod.yaml"
)

type Config struct {
	Mode     string
	DB       DBConfig      `yaml:"db" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTTL time.Duration `yaml:"tokenTTL" env-default:"1h"`
}

type DBConfig struct {
	Storage_path string `yaml:"storage_path" env-required:"true"`
}

type GRPCConfig struct {
	Url     string        `yaml:"url" env-required:"true"`
	Port    uint16        `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

// Загрузка конфигураций для приложения
func Load() *Config {
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

	return &cfg

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
