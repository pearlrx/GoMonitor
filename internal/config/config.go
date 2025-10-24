package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type DataBaseConfig struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Name       string `yaml:"name"`
	Migrations string `yaml:"migrations"`
}

type CollectorConfig struct {
	IntervalSeconds int `yaml:"intervalSeconds"`
}

type Server struct {
	Name        string `yaml:"name"`
	Address     string `yaml:"address"`
	Description string `yaml:"description"`
}

type Config struct {
	DataBase  DataBaseConfig  `yaml:"database"`
	Collector CollectorConfig `yaml:"collector"`
	Servers   []Server        `yaml:"servers"`
}

func Load(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config

	if err = yaml.Unmarshal(f, &cfg); err != nil {
		return nil, err
	}

	if cfg.Collector.IntervalSeconds <= 0 {
		cfg.Collector.IntervalSeconds = 10
	}
	return &cfg, nil
}
