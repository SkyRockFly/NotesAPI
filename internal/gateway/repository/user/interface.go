package userrepo

import (
	"context"
)

type User struct {
	ID       int
	Login    string
	Password string
	Email    string
}

type CreateUserReq struct {
	Login    string
	Password string
	Email    string
}

type IUser interface {
	Create(ctx context.Context, user CreateUserReq) (int, error)
	Delete(ctx context.Context, id int) error
	Get(ctx context.Context, login string) (User, error)
}
