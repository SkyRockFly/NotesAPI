package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	notesv1 "notes/internal/grpc/generated"
	noterepository "notes/internal/notes-svc/noteRepository"
	noteservice "notes/internal/notes-svc/noteService"
	"notes/internal/pkg/interceptors"
	"notes/internal/pkg/testutil"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const (
	resetFixtures = `TRUNCATE TABLE note RESTART IDENTITY CASCADE;`
	fixturePath   = "./testdata/fixtures/note.sql"
	bufSize       = 1024 * 1024
)

var (
	pool    *pgxpool.Pool
	svc     *noteservice.Service
	handler *Server
)

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

	handler = NewServer(svc)

	code := m.Run()
	os.Exit(code)
}

func ConfigureFromENV() (*TestCfg, error) {
	var cfg TestCfg
	cfg.DBurl = os.Getenv("NOTES_DB_URL")

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

func newTestGRPCClient(
	t *testing.T,
	handler notesv1.NotesServiceServer,
) notesv1.NotesServiceClient {
	t.Helper()

	listener := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.LogInterceptor(),
	))
	notesv1.RegisterNotesServiceServer(grpcServer, handler)

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Panicf("grpcServer serve: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(
			func(ctx context.Context, _ string) (net.Conn, error) {
				return listener.DialContext(ctx)
			},
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
	})

	return notesv1.NewNotesServiceClient(conn)
}
