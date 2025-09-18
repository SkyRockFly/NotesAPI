package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Log    LoggerConfig `yaml:"logger"`
}

func (a *AppConfig) Validate() error {
	if err := a.Server.validate(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	if err := a.Log.validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	if err := a.DB.validate(); err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

func (s *ServerConfig) validate() error {
	port, err := strconv.Atoi(s.Port)
	if err != nil {
		return fmt.Errorf(" port string convert:%w", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port out of range (1-65535):%v", port)
	}

	return nil
}

type DBConfig struct {
	URL string `yaml:"url"`
}

func (db *DBConfig) validate() error {
	if strings.TrimSpace(db.URL) == "" {
		return fmt.Errorf("empty url")
	}

	return nil
}

type LoggerConfig struct {
	Timestamp   string `yaml:"timestamp"`
	FormatLevel string `yaml:"formatlevel"`
	Level       string `yaml:"level"`
}

func (l *LoggerConfig) validate() error {
	_, err := time.Parse(time.DateTime, l.Timestamp)
	if err != nil {
		return fmt.Errorf("time parse: %w", err)
	}

	_, err = zerolog.ParseLevel(l.Level)
	if err != nil {
		return fmt.Errorf("level parse: %w", err)
	}

	return nil
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
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}
	return config, nil
}
