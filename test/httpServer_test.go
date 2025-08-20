package test

import (
	httpserver "NotesService/internal/httpServer"
	noterepository "NotesService/internal/noteRepository"
	noteservice "NotesService/internal/noteService"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	pool     *pgxpool.Pool
	db       *sql.DB
	fixtures *testfixtures.Loader
)

type fixtureNote struct {
	ID        int       `yaml:"id" json:"id"`
	AccountID int       `yaml:"account_id" json:"account_id"`
	Title     string    `yaml:"title" json:"title"`
	Body      string    `yaml:"body" json:"body"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

type fixtureTime struct {
	CreatedAt string `yaml:"created_at"`
	UpdatedAt string `yaml:"updated_at"`
}

type want struct {
	code int
	json map[string]any
}

type tc struct {
	name          string
	body          string
	jsonMediatype bool
	want          want
}

type reqInfo struct {
	method string
	url    string
}

func TestMain(m *testing.M) {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var dbURL string
	var pgC *postgres.PostgresContainer
	db, pgC, dbURL, err = initTestDB(cfg)
	if err != nil {
		log.Fatalf("initDB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer pgC.Terminate(ctx)
	defer db.Close()

	if err := gooseUp(db); err != nil {
		log.Fatalf("goose: %v", err)
	}

	pool, err = setupPgxPool(dbURL)
	if err != nil {
		log.Fatalf("pgxPool: %v", err)
	}
	defer pool.Close()

	fixtures, err = setupFixtures(db)
	if err != nil {
		log.Fatalf("fixtures: %v", err)
	}

	code := m.Run()
	os.Exit(code)

}

func TestFixturesLoaded(t *testing.T) {
	require.NoError(t, loadFixtures(), "load fixtures")
	var count int
	err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM note").Scan(&count)
	require.NoError(t, err)
	t.Logf("rows in note: %d", count)
	require.Greater(t, count, 0, "fixtures did not load into note table")
}

func testTransportValidation(t *testing.T, h http.HandlerFunc, r reqInfo) {
	cases := []tc{
		{
			name:          "Wrong mediatype",
			body:          `{"account_id":101,"title":"Honey","body":"Bears"}`,
			jsonMediatype: false,
			want: want{
				code: http.StatusUnsupportedMediaType,
				json: map[string]any{"error": "unsupported media type\n"},
			},
		},
		{
			name:          "Invalid json",
			body:          `{"account_id":101,`,
			jsonMediatype: true,
			want: want{
				code: http.StatusBadRequest,
				json: map[string]any{"error": "invalid json"},
			},
		},
		{
			name:          "Wrong fields",
			body:          `{"acc_id":101,"ttl":"x"}`,
			jsonMediatype: true,
			want: want{
				code: http.StatusBadRequest,
				json: map[string]any{"error": "invalid json"},
			},
		},
		{
			name:          "Invalid id",
			body:          `{"account_id":0,"title":"x","body":"y"}`,
			jsonMediatype: true,
			want: want{
				code: http.StatusBadRequest,
				json: map[string]any{"error": "invalid json"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			if c.jsonMediatype {
				req.Header.Set("Content-Type", "application/json")
			}

			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			if !c.jsonMediatype {
				require.Equal(t, c.want.json["error"], rr.Body.String(), "unsupported media type")
				return
			}

			var got map[string]any
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
			for k, v := range c.want.json {
				require.Equal(t, v, got[k], "json field %q", k)
			}

		})
	}
}

func TestHttpCreateNoteHandler(t *testing.T) {
	serv := setupService()
	h := httpserver.HTTPCreateNoteHandler(serv)

	cases := []tc{
		{
			name:          "Ok",
			body:          `{"account_id":101,"title":"Honey","body":"Bears"}`,
			jsonMediatype: true,
			want:          want{code: http.StatusCreated, json: nil},
		},
	}

	r := reqInfo{
		method: http.MethodPost,
		url:    "/notes/create",
	}
	testTransportValidation(t, h, r)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			req.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			if rr.Code >= 200 && rr.Code < 300 && c.want.json == nil {
				var resp struct {
					ID int `json:"id"`
				}
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				require.Greater(t, resp.ID, 0)
				return
			}
		})
	}

}

func TestHttpDeleteNoteHandler(t *testing.T) {
	if err := loadFixtures(); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	serv := setupService()
	h := httpserver.HTTPDeleteNoteHandler(serv)

	cases := []tc{
		{
			name: "Delete existing note",
			body: `{"account_id":303,"id":5}`,
			want: want{
				code: http.StatusOK,
				json: map[string]any{"deleted": true},
			},
		},
		{
			name: "Delete non-existing note",
			body: `{"account_id":101,"id":99999}`,
			want: want{
				code: http.StatusNotFound,
				json: map[string]any{"error": "note not found"},
			},
		},
	}

	r := reqInfo{
		method: http.MethodDelete,
		url:    "/notes/delete",
	}
	testTransportValidation(t, h, r)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			req.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			var got map[string]any
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))

			for k, v := range c.want.json {
				require.Equal(t, got[k], v)
			}

		})
	}

}

func TestHttpUpdateNoteHandler(t *testing.T) {
	if err := loadFixtures(); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	serv := setupService()
	h := httpserver.HTTPUpdateNoteHandler(serv)

	cases := []tc{
		{
			name: "Update existing note",
			body: `{"account_id":303,"id":5,"title":"Patch", "body":"I did it"}`,
			want: want{
				code: http.StatusOK,
				json: map[string]any{"updated": true},
			},
		},
		{
			name: "Update non-existing note",
			body: `{"account_id":101,"id":99999, "title":"Patch", "body":"I did it"}`,
			want: want{
				code: http.StatusNotFound,
				json: map[string]any{"error": "note not found"},
			},
		},
	}

	r := reqInfo{
		method: http.MethodPatch,
		url:    "/notes/update",
	}
	testTransportValidation(t, h, r)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			req.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			var got map[string]any
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))

			for k, v := range c.want.json {
				require.Equal(t, got[k], v)
			}

		})
	}

}

func TestHttpGetNoteHandler(t *testing.T) {
	if err := loadFixtures(); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	serv := setupService()
	h := httpserver.HTTPGetNoteHandler(serv)

	testNotes, err := loadFixtureNotes()
	require.NoError(t, err, "load test notes")

	testNote, err := chooseFixtureNoteId(testNotes, 3)
	require.NoError(t, err, "load test note")

	cases := []tc{
		{
			name: "Get existing note",
			body: `{"account_id":202,"id":3}`,
			want: want{
				code: http.StatusOK,
				json: nil,
			},
		},
		{
			name: "Get non-existing note",
			body: `{"account_id":101,"id":99999}`,
			want: want{
				code: http.StatusNotFound,
				json: map[string]any{"error": "note not found"},
			},
		},
	}

	r := reqInfo{
		method: http.MethodGet,
		url:    "/note/get",
	}
	testTransportValidation(t, h, r)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			req.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			if c.want.json != nil {
				var got map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				for k, v := range c.want.json {
					require.Equal(t, got[k], v)
				}
				return
			}

			var got fixtureNote
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
			compareNotes(t, testNote, got)

		})
	}

}

func TestHttpListNoteHandler(t *testing.T) {
	if err := loadFixtures(); err != nil {
		log.Fatalf("load fixtures: %v", err)
	}

	serv := setupService()
	h := httpserver.HTTPListNoteHandler(serv)

	test, err := loadFixtureNotes()
	require.NoError(t, err, "load test notes")

	testNotes, err := chooseFixtureNoteAccountID(test, 202)
	require.NoError(t, err, "choose accountID")

	cases := []tc{
		{
			name: "Get existing notes",
			body: `{"account_id":202}`,
			want: want{
				code: http.StatusOK,
				json: nil,
			},
		},
		{
			name: "Get non-existing notes",
			body: `{"account_id":1000}`,
			want: want{
				code: http.StatusNotFound,
				json: map[string]any{"error": "note not found"},
			},
		},
	}

	r := reqInfo{
		method: http.MethodGet,
		url:    "/notes/get",
	}
	testTransportValidation(t, h, r)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.url, strings.NewReader(c.body))
			rr := httptest.NewRecorder()

			req.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(rr, req)
			assert.Equal(t, c.want.code, rr.Code, "status code")

			if c.want.json != nil {
				var got map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				for k, v := range c.want.json {
					require.Equal(t, got[k], v)
				}
				return
			}

			var got []fixtureNote
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
			require.Equal(t, len(testNotes), len(got), "test length")

			sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
			sort.Slice(testNotes, func(i, j int) bool { return testNotes[i].ID < testNotes[j].ID })

			for i, note := range got {
				want := testNotes[i]
				compareNotes(t, note, want)
			}
		})
	}

}

func setupService() *noteservice.Service {
	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)
	return service
}

func loadFixtures() error {
	if err := fixtures.Load(); err != nil {
		return err
	}
	return nil
}

func loadFixtureNotes() ([]fixtureNote, error) {
	var notes []fixtureNote
	b, err := os.ReadFile("./fixtures/note.yml")
	if err != nil {
		return notes, err
	}
	if err := yaml.Unmarshal(b, &notes); err != nil {
		return notes, err
	}
	return notes, nil
}

func chooseFixtureNoteId(notes []fixtureNote, id int) (fixtureNote, error) {
	for _, note := range notes {
		if id == note.ID {
			return note, nil
		}
	}
	return fixtureNote{}, fmt.Errorf("fixtureNote not found")
}

func chooseFixtureNoteAccountID(notes []fixtureNote, id int) ([]fixtureNote, error) {
	fixtureNotes := make([]fixtureNote, 0, len(notes))
	for _, note := range notes {
		if id == note.AccountID {
			fixtureNotes = append(fixtureNotes, note)
		}
	}
	if len(fixtureNotes) == 0 {
		return fixtureNotes, fmt.Errorf("notes not found")
	}
	return fixtureNotes, nil
}

func compareNotes(t *testing.T, expected fixtureNote, actual fixtureNote) {
	assert.Equal(t, expected.ID, actual.ID, "id test")
	assert.Equal(t, expected.AccountID, actual.AccountID, "accountID test")
	assert.Equal(t, expected.Title, actual.Title, "title test")
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second*2, "createdAt test")
	if !assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second*2, "updatedAt test") {
		t.Fatalf("note comparison")
	}
}
