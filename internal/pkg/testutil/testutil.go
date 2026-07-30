package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NormalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
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

func LoadFixtures(pool *pgxpool.Pool, path string, resetFixtures string) error {
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
}
