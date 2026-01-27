package config

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `env:"APP_ENV" env-default:"local"`

	Server ServerConfig
	DB     DBConfig
}

type ServerConfig struct {
	Port         int           `env:"SERVER_PORT" env-required:"true"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-default:"10s"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     int    `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	DBName   string `env:"DB_NAME" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exsist " + path)
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failes to read config" + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("server", c.Server),
		slog.Any("db", c.DB),
	)
}
