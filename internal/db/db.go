package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultDBConnTimeout    = 20 * time.Second
	defaultPingRetryTimeout = 3 * time.Second
)

func InitDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	childCtx, cancel := context.WithTimeout(ctx, defaultDBConnTimeout)
	defer cancel()

	pgxpool, err := pgxpool.New(childCtx, url)
	if err != nil {
		return nil, fmt.Errorf("establish pgxpool connection: %w", err)
	}

	tryNum := 0
	for err := pgxpool.Ping(childCtx); err != nil; tryNum++ {
		if tryNum == 3 {
			return nil, fmt.Errorf("ping database: %w", err)
		}
		time.Sleep(defaultPingRetryTimeout)
		log.Printf("Ping database, try %d. Retry after %s", tryNum+1, defaultDBConnTimeout)
	}

	log.Printf("Successfully established connection")

	return pgxpool, nil
}
