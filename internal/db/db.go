package db

import (
	"context"
	"fmt"
	"log"
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
		log.Printf("Cannot ping database, retrying for 9 seconds")
		for i := 0; i < 3; i++ {
			err = pgxpool.Ping(childCtx)
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
		if err := pgxpool.Ping(childCtx); err != nil {
			return nil, fmt.Errorf("failed to ping database:%w", err)
		}

	}

	fmt.Println("Successfully established connection")

	return pgxpool, nil
}
