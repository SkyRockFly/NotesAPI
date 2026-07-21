package config

import (
	"bytes"
	"fmt"
	"net/url"
	"notes/internal/pkg/kit"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server      ServerConfig      `yaml:"server" validate:"required"`
	DB          DBConfig          `yaml:"db" validate:"required"`
	Log         LoggerConfig      `yaml:"logger" validate:"required"`
	NotesSvc    ConnectionConfig  `yaml:"notes-svc" validate:"required"`
	Auth        AuthConfig        `yaml:"auth" validate:"required"`
	RateLimiter RateLimiterConfig `yaml:"rate_limiter" validate:"required"`
}

type ServerConfig struct {
	Port int `yaml:"port" validate:"required,min=1,max=65535"`
}

type ConnectionConfig struct {
	IsGRPC bool   `yaml:"is_grpc"`
	Host   string `yaml:"host" validate:"required,hostname|ip"`
	Port   int    `yaml:"port" validate:"required,min=1,max=65535"`
}

type DBConfig struct {
	URL string `yaml:"url" validate:"required,dsn=postgres"`
}

type LoggerConfig struct {
	Timestamp   string `yaml:"timestamp" validate:"required,timestamp"`
	FormatLevel string `yaml:"formatlevel" validate:"required,formatlevel"`
	Level       string `yaml:"level" validate:"required,loglevel"`
}

type AuthConfig struct {
	JWTSecretFile string `yaml:"jwt_secret_file" validate:"required"`
	JWTSecret     []byte
}

type RateLimiterConfig struct {
	Requests        int           `yaml:"requests" validate:"min=1,max=10000"`
	Period          time.Duration `yaml:"period" validate:"min=1s,max=24h"`
	Burst           int           `yaml:"burst" validate:"min=1,max=50"`
	VisitorTTL      time.Duration `yaml:"visitor_ttl" validate:"min=1m,max=168h"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" validate:"min=1m,max=24h"`
}

func initValidator() (*validator.Validate, error) {
	v := validator.New()

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
		return nil, fmt.Errorf("validate struct: %w", err)
	}

	return config, nil
}

func GetSecretKey(auth *AuthConfig) error {
	b, err := os.ReadFile(auth.JWTSecretFile)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	key := bytes.TrimSpace(b)
	if len(key) < 32 {
		return fmt.Errorf("weak secret key")
	}
	auth.JWTSecret = key
	return nil
}
