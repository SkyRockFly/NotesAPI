package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	pgxpool, err := pgxpool.New(childCtx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to establish pgxpool connection:%w", err)
	}

	err = pgxpool.Ping(childCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database:%w", err)
	}

	fmt.Println("Succesfully established connection")

	return pgxpool, nil
}
