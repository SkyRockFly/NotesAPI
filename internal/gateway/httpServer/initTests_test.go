package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	noterepo "notes/internal/gateway/repository/note"
	notehttp "notes/internal/gateway/repository/note/http"
	refreshtokenpg "notes/internal/gateway/repository/refreshToken/pg"
	userpg "notes/internal/gateway/repository/user/pg"
	authsvc "notes/internal/gateway/service/auth"
	"notes/internal/gateway/service/refreshsvc"
	usersvc "notes/internal/gateway/service/user"
	"notes/internal/pkg/testutil"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	pool    *pgxpool.Pool
	svcAuth *authsvc.Service
)

type TestCfg struct {
	DBurl string `yaml:"dbURL"`
}

func ConfigureFromENV() (*TestCfg, error) {
	var cfg TestCfg
	cfg.DBurl = os.Getenv("DB_URL")
	return &cfg, nil
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

	pgToken := refreshtokenpg.NewRepository(pool)
	svcToken := refreshsvc.NewService(pgToken, []byte("le hard secret"))

	pgUser := userpg.NewRepository(pool)
	svcUser := usersvc.NewService(pgUser)

	svcAuth = authsvc.New(svcToken, svcUser)

	code := m.Run()
	os.Exit(code)
}

func launchNotesServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/note/create":
			var req noterepo.CreateReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("decode create request: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch req.AccountID {
			case 101:
				require.Equal(t, "test title", req.Title)
				require.Equal(t, "test body", req.Body)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1}`))
			case 102:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"not found"}`))

			default:
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"service error"}`))
			}

		case r.Method == http.MethodPost && r.URL.Path == "/note/get":
			var req noterepo.GetReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("decode create request: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch req.ID {
			case 1:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
			"id": 1,
			"account_id": 101,
			"title": "test title",
			"body": "test body",
			"created_at": "2026-07-13T20:00:00Z",
			"updated_at": "2026-07-13T20:00:00Z"
		}`))
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"not found"}`))
			}

		case r.Method == http.MethodPost && r.URL.Path == "/notes/get":
			var req noterepo.ListReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("decode create request: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if req.Cursor < 0 || req.Limit < 0 {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			switch req.AccountID {
			case 101:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"cursor_next":2,"cursor_prev":1,"has_more":false,"notes":[
				{
					"id": 1,
			"account_id": 101,
			"title": "test title",
			"body": "test body",
			"created_at": "2026-07-13T20:00:00Z",
			"updated_at": "2026-07-13T20:00:00Z"
				},
				{
				"id": 2,
			"account_id": 101,
			"title": "test title",
			"body": "test body",
			"created_at": "2026-07-13T20:00:00Z",
			"updated_at": "2026-07-13T20:00:00Z"
				}
			]}`))
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"not found"}`))
			}

		case r.Method == http.MethodPatch && r.URL.Path == "/note/update":
			var req noterepo.UpdateReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode update request: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch req.AccountID {
			case 101:
				switch req.ID {
				case 1:
					require.Equal(t, "title update", req.Title)
					require.Equal(t, "body update", req.Body)
					w.WriteHeader(http.StatusNoContent)
					_, _ = w.Write([]byte(""))
				case 2:
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"not found"}`))

				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"service error"}`))
				}
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"not found"}`))
			}

		case r.Method == http.MethodDelete && r.URL.Path == "/note/delete":
			var req noterepo.DeleteReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode delete request: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			switch req.ID {
			case 1:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"deleted":true}`))

			case 2:
				{
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			default:
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"service error"}`))
			}

		default:
			http.NotFound(w, r)
		}
	}))
}

func newTestRepo(t *testing.T) *notehttp.HTTPRepo {
	t.Helper()

	srv := launchNotesServer(t)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)

	noteCFG := notehttp.SvcHTTPCfg{
		Host: u.Hostname(),
		Port: port,
	}

	return notehttp.NewHTTPRepo(noteCFG)
}
