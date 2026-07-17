package config

import (
	"fmt"
	"net/url"
	"notes/internal/pkg/kit"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func initValidator() (*validator.Validate, error) {
	v := validator.New()

	if err := v.RegisterValidation("port", validatePort); err != nil {
		return nil, fmt.Errorf("validate port: %w", err)
	}

	if err := v.RegisterValidation("dsn", validateDSN); err != nil {
		return nil, fmt.Errorf("validate dsn: %w", err)
	}

	if err := v.RegisterValidation("loglevel", validateZerologLevel); err != nil {
		return nil, fmt.Errorf("validate logLevel: %w", err)
	}

	if err := v.RegisterValidation("formatlevel", validateFormatLevel); err != nil {
		return nil, fmt.Errorf("validate formatLevel: %w", err)
	}

	if err := v.RegisterValidation("timestamp", validateTimestamp); err != nil {
		return nil, fmt.Errorf("validate timestamp: %w", err)
	}

	return v, nil
}

func validatePort(fl validator.FieldLevel) bool {
	s := strings.TrimSpace(fl.Field().String())
	port, err := strconv.Atoi(s)
	return err == nil && port >= 1 && port <= 65535
}

func validateZerologLevel(fl validator.FieldLevel) bool {
	s := strings.TrimSpace(fl.Field().String())
	_, err := zerolog.ParseLevel(s)
	return err == nil
}

func validateDSN(fl validator.FieldLevel) bool {
	s := strings.TrimSpace(fl.Field().String())
	if s == "" {
		log.Error().Msg("empty string")
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Error().Err(fmt.Errorf("parse: %w", err)).
			Str("scheme", u.Scheme).
			Str("host", u.Host).
			Msg("bad url")
		return false
	}

	param := fl.Param()
	if param == "" {
		param = "postgres"
	}

	allowed := map[string]bool{}

	for sch := range strings.SplitSeq(param, "|") {
		allowed[strings.ToLower(strings.TrimSpace(sch))] = true
	}

	return allowed[strings.ToLower(u.Scheme)]
}

func validateFormatLevel(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	defer func() { _ = recover() }()
	out := fmt.Sprintf(s, "info")
	return !strings.Contains(out, "%!")
}

func validateTimestamp(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	return strings.Contains(s, "2006")
}

type AppConfig struct {
	Server ServerConfig `yaml:"server" validate:"required"`
	DB     DBConfig     `yaml:"db" validate:"required"`
	Log    LoggerConfig `yaml:"logger" validate:"required"`
}

type ServerConfig struct {
	Port string `yaml:"port" validate:"required,port"`
}

type DBConfig struct {
	URL string `yaml:"url" validate:"required,dsn=postgres"`
}

type LoggerConfig struct {
	Timestamp   string `yaml:"timestamp" validate:"required,timestamp"`
	FormatLevel string `yaml:"formatlevel" validate:"required,formatlevel"`
	Level       string `yaml:"level" validate:"required,loglevel"`
}

func GetAppConfig() (*AppConfig, error) {
	config := &AppConfig{}

	v, err := initValidator()
	if err != nil {
		return nil, fmt.Errorf("GetConfigApp: %w", err)
	}

	data, err := os.ReadFile("./configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("GetAppConfig: %w", err)
	}

	if err := kit.ValidateStruct(v, config); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
