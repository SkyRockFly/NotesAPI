package httpserver

import (
	noteservice "NotesService/internal/noteService"
	"NotesService/internal/testutil"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool      *pgxpool.Pool
	svc       *noteservice.Service
	testNotes []FixtureNote
	testNote  FixtureNote
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

	testNotes, err = loadAccountNotes(cfg.FixtureAccID)
	if err != nil {
		log.Fatalf("load fixtureNotes: %v", err)
	}

	testNote, err = takeNotesFromFixtureList(testNotes, cfg.FixtureNoteID)
	if err != nil {
		log.Fatalf("load fixtureNote: %v", err)
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

func loadAccountNotes(accountID int) ([]FixtureNote, error) {
	rawList, err := svc.List(context.Background(),
		noteservice.Note{
			AccountID: accountID,
		})
	if err != nil {
		return nil, fmt.Errorf("take list: %w", err)
	}

	list := make([]FixtureNote, 0, len(rawList))

	var fixtureNote FixtureNote
	for _, note := range rawList {
		fixtureNote.ID = note.ID
		fixtureNote.AccountID = note.AccountID
		fixtureNote.Title = note.Title
		fixtureNote.Body = note.Body
		fixtureNote.CreatedAt = note.CreatedAt
		fixtureNote.UpdatedAt = note.UpdatedAt

		list = append(list, fixtureNote)

	}

	return list, nil
}

func takeNotesFromFixtureList(fixtureList []FixtureNote, noteID int) (FixtureNote, error) {
	for _, note := range fixtureList {
		if note.ID == noteID {
			return note, nil
		}
	}
	return FixtureNote{}, fmt.Errorf("no note found")

}

func normalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}
