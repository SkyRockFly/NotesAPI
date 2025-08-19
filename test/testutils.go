package test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/goccy/go-yaml"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testCfg struct {
	DB struct {
		URL      string            `yaml:"url"`
		Image    string            `yaml:"image"`
		Name     string            `yaml:"name"`
		User     string            `yaml:"user"`
		Password string            `yaml:"password"`
		Driver   string            `yaml:"driver"`
		Port     string            `yaml:"port"`
		Args     map[string]string `yaml:"args"`
	} `yaml:"db"`
}

func loadConfig() (*testCfg, error) {
	cfg := &testCfg{}
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return cfg, fmt.Errorf("read: %w", err)
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal: %w", err)
	}

	return cfg, nil
}

func initTestDB(cfg *testCfg) (*sql.DB, *postgres.PostgresContainer, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pgC, err := postgres.Run(ctx,
		cfg.DB.Image,
		postgres.WithDatabase(cfg.DB.Name),
		postgres.WithUsername(cfg.DB.User),
		postgres.WithPassword(cfg.DB.Password),
		testcontainers.WithWaitStrategy(
			wait.ForSQL(nat.Port(cfg.DB.Port), cfg.DB.Driver,
				func(host string, port nat.Port) string {
					return pgURL(cfg, host, port)
				},
			).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, nil, "", fmt.Errorf("run db:%w", err)
	}

	dbURL, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, "", fmt.Errorf("create link: %w", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("load db: %w", err)
	}

	tryNum := 0
	for err := db.PingContext(ctx); err != nil; tryNum++ {
		if tryNum == 3 {
			return nil, nil, "", fmt.Errorf("ping db: %w", err)
		}
		time.Sleep(time.Second * 10)
	}

	return db, pgC, dbURL, nil
}

func pgURL(cfg *testCfg, host string, port nat.Port) string {
	u := &url.URL{
		Scheme: cfg.DB.Driver,
		User:   url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:   net.JoinHostPort(host, port.Port()),
		Path:   cfg.DB.Name,
	}

	if len(cfg.DB.Args) > 0 {
		q := url.Values{}
		for k, v := range cfg.DB.Args {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func gooseUp(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..")
	migrations := filepath.Join(root, "migrations")

	if err := goose.Up(db, migrations); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	return nil
}

func setupPgxPool(dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool create: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return pool, nil
}

func setupFixtures(db *sql.DB) (*testfixtures.Loader, error) {
	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures"),
	)
	if err != nil {
		return nil, fmt.Errorf("fixtures init: %w", err)
	}
	return fixtures, nil
}
