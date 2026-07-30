package refreshsvc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	refreshtokenrepo "notes/internal/gateway/repository/refreshToken"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/kit"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	shortTTL      = time.Hour
	longTTL       = 30 * 24 * time.Hour
	longThreshold = shortTTL + (longTTL-shortTTL)/2
	// Rotate a long-lived refresh token when its remaining TTL
	// is less than or equal to this threshold.
	rotateThreshold = 20 * 24 * time.Hour

	jwtTTL = 10 * time.Minute
)

type Service struct {
	repo     refreshtokenrepo.IToken
	secret   []byte
	validate *validator.Validate
}

type IssueTokensReq struct {
	UserID     int `validate:"min=1"`
	RememberMe bool
}

type RevokeTokenReq struct {
	Refresh string
}

type RevokeAllReq struct {
	UserID int `validate:"min=1"`
}

type refreshKey struct {
	selector string
	private  []byte
}

type AuthResp struct {
	Access           string
	Refresh          string
	RefreshExpiresAt time.Time
}

func NewService(r refreshtokenrepo.IToken, secret []byte) *Service {
	return &Service{
		repo:     r,
		secret:   secret,
		validate: validator.New(),
	}
}

func (s *Service) Extend(ctx context.Context, refresh string) (AuthResp, error) {
	oldSelector, oldPrivate, err := parseRefresh(refresh)
	if err != nil {
		return AuthResp{}, fmt.Errorf("parse refresh: %w", err)
	}

	oldToken, err := s.repo.Get(ctx, oldSelector)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return AuthResp{}, fmt.Errorf("get token: %w", apperror.ErrUnauthorized)
		}
		return AuthResp{}, fmt.Errorf("get token: %w", err)
	}

	if oldToken.Revoked || time.Now().After(oldToken.ExpiredAt) {
		return AuthResp{}, fmt.Errorf("%w, expired or revoked", apperror.ErrUnauthorized)
	}

	if subtle.ConstantTimeCompare(oldToken.PrivateHash, sha256Sum(oldPrivate)) != 1 {
		return AuthResp{}, fmt.Errorf("%w, bad hash", apperror.ErrUnauthorized)
	}

	jwt, err := generateJWTToken(s.secret, oldToken.UserID)
	if err != nil {
		return AuthResp{}, fmt.Errorf("grant jwt: %w", err)
	}

	resp := AuthResp{
		Access: jwt,
	}

	isLong := oldToken.ExpiredAt.Sub(oldToken.IssuedAt) > longThreshold
	isExpired := time.Until(oldToken.ExpiredAt) <= rotateThreshold

	if isLong && isExpired {
		key, err := generateRefreshToken()
		if err != nil {
			return AuthResp{}, fmt.Errorf("generate refresh: %w", err)
		}

		token := refreshtokenrepo.CreateReq{
			Selector: key.selector,
			Private:  sha256Sum(key.private),
			UserID:   oldToken.UserID,
			TTL:      longTTL,
		}
		if err := s.repo.Create(ctx, token); err != nil {
			return AuthResp{}, fmt.Errorf("create refresh:%w", err)
		}

		if err := s.repo.Revoke(ctx, oldSelector); err != nil {
			return AuthResp{}, fmt.Errorf("revoke old: %w", err)
		}

		resp.Refresh = key.pairKeys()
		resp.RefreshExpiresAt = time.Now().Add(longTTL)
	}

	return resp, nil
}

func (s *Service) IssueTokens(ctx context.Context, req IssueTokensReq) (AuthResp, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return AuthResp{}, fmt.Errorf("%w: validate request struct: %w", apperror.ErrBadRequest, err)
	}

	jwt, err := generateJWTToken(s.secret, req.UserID)
	if err != nil {
		return AuthResp{}, fmt.Errorf("generate jwt token: %w", err)
	}

	key, err := generateRefreshToken()
	if err != nil {
		return AuthResp{}, fmt.Errorf("generate refresh token: %w", err)
	}

	ttl := shortTTL
	if req.RememberMe {
		ttl = longTTL
	}

	refToken := refreshtokenrepo.CreateReq{
		Selector: key.selector,
		Private:  sha256Sum(key.private),
		UserID:   req.UserID,
		TTL:      ttl,
	}
	if err := s.repo.Create(ctx, refToken); err != nil {
		return AuthResp{}, fmt.Errorf("create refresh token: %w", err)
	}

	resp := AuthResp{
		Access:           jwt,
		Refresh:          key.pairKeys(),
		RefreshExpiresAt: time.Now().Add(ttl),
	}

	return resp, nil
}

func (s *Service) RevokeAll(ctx context.Context, req RevokeAllReq) error {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return fmt.Errorf("%w: validate request struct: %w", apperror.ErrBadRequest, err)
	}

	if err := s.repo.RevokeAll(ctx, req.UserID); err != nil {
		return fmt.Errorf("revokeAll: %w", err)
	}

	return nil
}

func (s *Service) RevokeToken(ctx context.Context, req RevokeTokenReq) error {
	oldSelector, _, err := parseRefresh(req.Refresh)
	if err != nil {
		return fmt.Errorf("parse refresh: %w", err)
	}

	if err := s.repo.Revoke(ctx, oldSelector); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}

func generateRefreshToken() (refreshKey, error) {
	selector := uuid.New()

	private := make([]byte, 32)
	if _, err := rand.Read(private); err != nil {
		return refreshKey{}, fmt.Errorf("generate refresh private: %w", err)
	}

	return refreshKey{
		selector: selector.String(),
		private:  private,
	}, nil
}

func generateJWTToken(secret []byte, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(int64(id), 10),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(jwtTTL)),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return tokenString, nil
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

func parseRefresh(refresh string) (string, []byte, error) {
	parts := strings.Split(refresh, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("%w: invalid refresh format", apperror.ErrBadRequest)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("%w: invalid private part", apperror.ErrBadRequest)
	}

	return parts[0], raw, nil
}

func (k refreshKey) pairKeys() string {
	return k.selector + "." + base64.RawURLEncoding.EncodeToString(k.private)
}
