package authsvc

import (
	"context"
	"fmt"
	"notes/internal/gateway/service/refreshsvc"
	usersvc "notes/internal/gateway/service/user"
)

type IAuthSVC interface {
	Login(ctx context.Context, req LoginReq) (AuthResp, error)
	Signup(ctx context.Context, req SignUpReq) (AuthResp, error)
	Refresh(ctx context.Context, refresh string) (AuthResp, error)
}

type LoginReq struct {
	Login      string
	Password   string
	RememberMe bool
}

type SignUpReq struct {
	Login    string
	Password string
	Email    string
}

type AuthResp struct {
	Access  string
	Refresh string
}

type Service struct {
	tokenSVC *refreshsvc.Service
	userSVC  *usersvc.Service
}

func New(token *refreshsvc.Service, user *usersvc.Service) *Service {
	return &Service{
		tokenSVC: token,
		userSVC:  user,
	}
}

func (s *Service) Login(ctx context.Context, req LoginReq) (AuthResp, error) {
	user := usersvc.GetUserReq{
		Login:    req.Login,
		Password: req.Password,
	}

	id, err := s.userSVC.Get(ctx, user)
	if err != nil {
		return AuthResp{}, fmt.Errorf("get user: %w", err)
	}

	tokenReq := refreshsvc.IssueTokensReq{
		UserID:     id,
		RememberMe: req.RememberMe,
	}

	tokens, err := s.tokenSVC.IssueTokens(ctx, tokenReq)
	if err != nil {
		return AuthResp{}, fmt.Errorf("authorize user:%w", err)
	}

	resp := AuthResp{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}

	return resp, nil
}

func (s *Service) Signup(ctx context.Context, req SignUpReq) (AuthResp, error) {
	user := usersvc.CreateUserReq{
		Login:    req.Login,
		Password: req.Password,
		Email:    req.Email,
	}

	id, err := s.userSVC.Create(ctx, user)
	if err != nil {
		return AuthResp{}, fmt.Errorf("create user: %w", err)
	}

	tokenReq := refreshsvc.IssueTokensReq{
		UserID:     id,
		RememberMe: true,
	}

	tokens, err := s.tokenSVC.IssueTokens(ctx, tokenReq)
	if err != nil {
		return AuthResp{}, fmt.Errorf("authorize user: %w", err)
	}

	resp := AuthResp{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}
	return resp, nil
}

func (s *Service) Refresh(ctx context.Context, refresh string) (AuthResp, error) {
	tokens, err := s.tokenSVC.Extend(ctx, refresh)
	if err != nil {
		return AuthResp{}, fmt.Errorf("%w", err)
	}
	resp := AuthResp{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}
	return resp, nil
}
