package httpserver

import (
	"fmt"
	"log"
	noterepository "notes/internal/notes-svc/noteRepository"
	noteservice "notes/internal/notes-svc/noteService"
	"notes/internal/pkg/testutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert/yaml"
)

const (
	resetFixtures = `TRUNCATE TABLE note RESTART IDENTITY CASCADE;`
	fixturePath   = "./testdata/fixtures/note.sql"
)

var (
	pool *pgxpool.Pool
	svc  *noteservice.Service
)

type FixtureNote struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TestCfg struct {
	DBurl         string `yaml:"dbURL"`
	FixtureAccID  int    `yaml:"fixture_account_id"`
	FixtureNoteID int    `yaml:"fixture_note_id"`
}

func TestMain(m *testing.M) {
	cfg, err := ConfigureFromENV()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err = testutil.SetupPgxPool(cfg.DBurl)
	if err != nil {
		log.Fatalf("pgxPool: %v", err)
	}
	defer pool.Close()

	svc = SetupRealService(pool)
	if err := testutil.LoadFixtures(pool, fixturePath, resetFixtures); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	code := m.Run()
	os.Exit(code)

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

func SetupRealService(pool *pgxpool.Pool) *noteservice.Service {
	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)
	return service
}
