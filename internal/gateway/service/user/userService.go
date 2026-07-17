package usersvc

import (
	"context"
	"errors"
	"fmt"
	userrepo "notes/internal/gateway/repository/user"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/kit"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

/*
	using to counter timing attacks by masking time difference between "user not found" and "invalid password",

so, it is used to launch bcrypt.CompareHashAndPassword if get user returned error
*/
var dummyHash = []byte("$2a$12$Vm3j3gU1dY3kB5jfiihvzeWYk6Jy6iitadj4a5ZMmJ5aO/cNVqEuC")

const (
	bCryptCost = 12
)

type Service struct {
	repo     userrepo.IUser
	validate *validator.Validate
}

type CreateUserReq struct {
	Login    string `validate:"required,min=1,max=255"`
	Password string `validate:"required,min=8,max=71"`
	Email    string `validate:"required,email"`
}

type GetUserReq struct {
	Login    string `validate:"required,min=1,max=255"`
	Password string `validate:"required,min=8,max=71"`
}

func NewService(r userrepo.IUser) *Service {
	return &Service{
		repo:     r,
		validate: validator.New(),
	}
}

func (s *Service) Create(ctx context.Context, req CreateUserReq) (int, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return 0, fmt.Errorf("validate req: %w", apperror.ErrBadRequest)
	}

	if len(req.Password) > 72 {
		return 0, fmt.Errorf("%w: password more than 72 bytes", apperror.ErrBadRequest)
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bCryptCost)
	if err != nil {
		return 0, fmt.Errorf("generate hash: %w", err)
	}

	user := userrepo.CreateUserReq{
		Login:    req.Login,
		Password: string(passHash),
		Email:    req.Email,
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if id < 1 {
		return fmt.Errorf("%w: id less than one", apperror.ErrBadRequest)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

func (s *Service) Get(ctx context.Context, req GetUserReq) (int, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return 0, fmt.Errorf("%w,validate struct: %w", apperror.ErrBadRequest, err)
	}

	user, err := s.repo.Get(ctx, req.Login)
	if err != nil {
		// using dummy hash to mask time difference between "user not found" and "invalid password"
		if errors.Is(err, apperror.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(
				dummyHash,
				[]byte(req.Password),
			)

			return 0, fmt.Errorf("unauthorized: %w", apperror.ErrUnauthorized)
		}
		return 0, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return 0, fmt.Errorf("%w,compare password: %w", apperror.ErrUnauthorized, err)
	}

	return user.ID, nil
}
