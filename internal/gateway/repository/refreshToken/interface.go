package refreshtokenrepo

import (
	"context"
	"time"
)

type Token struct {
	Selector    string
	PrivateHash []byte
	UserID      int
	IssuedAt    time.Time
	ExpiredAt   time.Time
	Revoked     bool
}

type CreateReq struct {
	Selector string
	Private  []byte
	UserID   int
	TTL      time.Duration
}

type IToken interface {
	Create(ctx context.Context, req CreateReq) error
	Revoke(ctx context.Context, selector string) error
	RevokeAll(ctx context.Context, userID int) error
	Get(ctx context.Context, selector string) (Token, error)
}
