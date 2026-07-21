package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	noterepo "notes/internal/gateway/repository/note"
	notehttp "notes/internal/gateway/repository/note/http"
	notemock "notes/internal/gateway/repository/note/mock"
	refreshtokenpg "notes/internal/gateway/repository/refreshToken/pg"
	userpg "notes/internal/gateway/repository/user/pg"
	authsvc "notes/internal/gateway/service/auth"
	"notes/internal/gateway/service/refreshsvc"
	usersvc "notes/internal/gateway/service/user"
	notesv1 "notes/internal/grpc/generated"
	"notes/internal/pkg/interceptors"
	"notes/internal/pkg/testutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1024 * 1024

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
				w.WriteHeader(http.StatusNoContent)
				_, _ = w.Write([]byte(""))

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

func newTestGRPCServer(
	t *testing.T,
) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.LogInterceptor(),
	))

	mock := createFakeServer(t)

	notesv1.RegisterNotesServiceServer(grpcServer, &mock)

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Panicf("grpcServer serve: %v", err)
		}
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		_ = listener.Close()
	})

	return listener.Addr().String()
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

func createFakeServer(t *testing.T) notemock.FakeNotesServer {
	mock := notemock.FakeNotesServer{
		CreateNoteFn: func(ctx context.Context,
			req *notesv1.CreateNoteReq,
		) (*notesv1.CreateNoteResp, error) {
			switch req.GetAccountId() {
			case 101:
				assert.Equal(t, "test title", req.Title)
				assert.Equal(t, "test body", req.Body)
				resp := &notesv1.CreateNoteResp{
					Id: 1,
				}
				return resp, nil
			case 102:
				return &notesv1.CreateNoteResp{}, status.Error(codes.NotFound, "not found")

			default:
				return nil, status.Error(codes.Unimplemented, "not implemented")
			}
		},
		GetNoteFn: func(
			ctx context.Context,
			req *notesv1.GetNoteReq,
		) (*notesv1.GetNoteResp, error) {
			switch req.GetId() {
			case 1:
				assert.Equal(t, int32(101), req.GetAccountId())
				return &notesv1.GetNoteResp{
					Id:        1,
					AccountId: 101,
					Title:     "test title",
					Body:      "test body",
					CreatedAt: timestamppb.New(
						time.Date(
							2026, time.July, 13,
							20, 0, 0, 0,
							time.UTC,
						),
					),
					UpdatedAt: timestamppb.New(
						time.Date(
							2026, time.July, 13,
							20, 0, 0, 0,
							time.UTC,
						),
					),
				}, nil

			default:
				return nil, status.Error(
					codes.NotFound,
					"not found",
				)
			}
		},
		ListNoteFn: func(
			ctx context.Context,
			req *notesv1.ListNoteReq,
		) (*notesv1.ListNoteResp, error) {
			if req.GetCursor() < 0 || req.GetLimit() < 0 {
				return nil, status.Error(
					codes.InvalidArgument,
					"bad request",
				)
			}

			switch req.GetAccountId() {
			case 101:
				createdAt := timestamppb.New(
					time.Date(
						2026, time.July, 13,
						20, 0, 0, 0,
						time.UTC,
					),
				)

				updatedAt := timestamppb.New(
					time.Date(
						2026, time.July, 13,
						20, 0, 0, 0,
						time.UTC,
					),
				)

				return &notesv1.ListNoteResp{
					CursorNext: 2,
					CursorPrev: 1,
					HasMore:    false,
					Notes: []*notesv1.Note{
						{
							Id:        1,
							AccountId: 101,
							Title:     "test title",
							Body:      "test body",
							CreatedAt: createdAt,
							UpdatedAt: updatedAt,
						},
						{
							Id:        2,
							AccountId: 101,
							Title:     "test title",
							Body:      "test body",
							CreatedAt: createdAt,
							UpdatedAt: updatedAt,
						},
					},
				}, nil

			default:
				return nil, status.Error(
					codes.NotFound,
					"not found",
				)
			}
		},
		UpdateNoteFn: func(
			ctx context.Context,
			req *notesv1.UpdateNoteReq,
		) (*notesv1.UpdateNoteResp, error) {
			switch req.GetAccountId() {
			case 101:
				switch req.GetId() {
				case 1:
					assert.Equal(t, "title update", req.GetTitle())
					assert.Equal(t, "body update", req.GetBody())

					return &notesv1.UpdateNoteResp{}, nil

				case 2:
					return nil, status.Error(
						codes.NotFound,
						"not found",
					)

				default:
					return nil, status.Error(
						codes.Internal,
						"service error",
					)
				}

			default:
				return nil, status.Error(
					codes.NotFound,
					"not found",
				)
			}
		},

		DeleteNoteFn: func(
			ctx context.Context,
			req *notesv1.DeleteNoteReq,
		) (*notesv1.DeleteNoteResp, error) {
			switch req.GetId() {
			case 1:
				return &notesv1.DeleteNoteResp{}, nil

			case 2:
				return nil, status.Error(
					codes.NotFound,
					"not found",
				)

			default:
				return nil, status.Error(
					codes.Internal,
					"service error",
				)
			}
		},
	}

	return mock
}

func makeGRPCClient(address string) (notesv1.NotesServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("new grpc client: %w", err)
	}
	client := notesv1.NewNotesServiceClient(conn)
	return client, conn, nil
}
