package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env_default:"dev"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Storage    Storage    `yaml:"storage"`
	Kafka      Kafka      `yaml:"kafka"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Storage struct {
	Postgres struct {
		Dsn          string `yaml:"dsn"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxIdleTime  string `yaml:"max_idle_time"`
	} `yaml:"postgres"`
	LruCache struct {
		Capacity int `yaml:"capacity"`
	} `yaml:"lru_cache"`
}

type Kafka struct {
	Addresses []string `yaml:"addresses"`

	Consumer struct {
		OrderTopic string `yaml:"order_topic"`
		OrderGroup string `yaml:"order_group"`
	} `yaml:"consumer"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Load config path is failed")
	}

	//panic("Load config failed")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exists", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Parse config is failed: %s", err.Error())
	}
	return &cfg
}
