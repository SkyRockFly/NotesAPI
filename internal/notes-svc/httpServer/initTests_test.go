package httpserver

import (
	"bytes"
	"encoding/json"
	"log"
	noteservice "notes/internal/notes-svc/noteService"
	"notes/internal/notes-svc/testutil"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	svc  *noteservice.Service
)

func TestMain(m *testing.M) {
	cfg, err := testutil.ConfigureFromENV()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err = testutil.SetupPgxPool(cfg.DBurl)
	if err != nil {
		log.Fatalf("pgxPool: %v", err)
	}
	defer pool.Close()

	svc = testutil.SetupRealService(pool)

	if err := testutil.LoadFixtures(pool, fixturePath); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	code := m.Run()
	os.Exit(code)

}

type FixtureNote struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func normalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}
