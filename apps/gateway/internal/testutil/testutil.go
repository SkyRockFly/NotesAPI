package testutil

/* import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

const (
	resetFixtures = `TRUNCATE TABLE note RESTART IDENTITY CASCADE;`
)

type TestCfg struct {
	DBurl         string `yaml:"dbURL"`
	FixtureAccID  int    `yaml:"fixture_account_id"`
	FixtureNoteID int    `yaml:"fixture_note_id"`
}

func ConfigureFromENV() (*TestCfg, error) {
	var cfg TestCfg
	cfg.DBurl = os.Getenv("DB_URL")
	fixtureAccID := os.Getenv("FIXTURE_ACCOUNT_ID")

	var err error
	cfg.FixtureAccID, err = strconv.Atoi(fixtureAccID)
	if err != nil {
		return &cfg, fmt.Errorf("convert fixture acc ID: %w", err)
	}
	if cfg.FixtureAccID < 0 {
		return &cfg, fmt.Errorf("fixture account id less than zero")
	}

	fixtureNoteID := os.Getenv("FIXTURE_NOTE_ID")
	cfg.FixtureNoteID, err = strconv.Atoi(fixtureNoteID)
	if err != nil {
		return &cfg, fmt.Errorf("convert fixture acc ID: %w", err)
	}

	if cfg.FixtureNoteID < 0 {
		return &cfg, fmt.Errorf("fixture account id less than zero")
	}
	return &cfg, nil
}

func ConfigureFromYAML(path string) (TestCfg, error) {
	var cfg TestCfg

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("readfile: %w", err)
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return cfg, nil
}

func SetupPgxPool(dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool create: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return pool, nil
}

func SetupRealService(pool *pgxpool.Pool) *noteservice.Service {
	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)
	return service
}

func LoadFixtures(pool *pgxpool.Pool, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	_, err = pool.Exec(context.Background(), resetFixtures)
	if err != nil {
		return fmt.Errorf("exec reset fixtures:%w", err)
	}

	_, err = pool.Exec(context.Background(), string(data))
	if err != nil {
		return fmt.Errorf("exec fixture: %w", err)
	}

	return nil
} */
