package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// для каждой структуры свой метод validate
type AppConfig struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Log    LoggerConfig `yaml:"logger"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DBConfig struct {
	URL string `yaml:"url"`
}

type LoggerConfig struct {
	Timestamp   string `yaml:"timestamp"`
	FormatLevel string `yaml:"formatlevel"`
	Level       string `yaml:"level"`
}

func GetAppConfig() (*AppConfig, error) {
	config := &AppConfig{}
	data, err := os.ReadFile("./configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}
	return config, nil
}

func validate(cfg *AppConfig) error {
	if cfg.DB.URL == "" {
		return fmt.Errorf("empty DB URL")
	}
	if cfg.Server.Port == "" {
		return fmt.Errorf("empty server port ")
	}

	port, err := strconv.Atoi(cfg.Server.Port)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port out of range (1-65535):%v", port)
	}
	return nil
}
