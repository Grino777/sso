package config

type Config struct {
	db   DbConf
	grpc GRPCConf
}

type DbConf struct {
	storage_path string `yaml:"storage_path" env-required:"true"`
}

type GRPCConf struct {
	url  string `yaml:"url" env-required:"true"`
	port uint16 `yaml:"port" env-required:"true"`
}

// func MustLoad() *Config {
// 	cfg :=
// }

// func MustLoadPath() {
// 	if _, err := os.Stat(name string)
// }
