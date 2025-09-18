package usernoterepo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")   // 404
	ErrBadRequest = errors.New("bad request") //400
	ErrUpstream   = errors.New("upstream")    //500
)

type httpRepo struct {
	base   string
	client *http.Client
}

func NewHttpRepo(base string) *httpRepo {
	return &httpRepo{
		base:   base,
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *httpRepo) Get(ctx context.Context, id int, accountID int) (UserNote, error) {
	req := struct {
		ID        int `json:"id"`
		AccountID int `json:"account_id"`
	}{id, accountID}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&req); err != nil {
		return UserNote{}, fmt.Errorf("encode: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		s.base+"note/get", &buf)
	if err != nil {
		return UserNote{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return UserNote{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var note struct {
			ID        int       `json:"id"`
			AccountID int       `json:"account_id"`
			Title     string    `json:"title"`
			Body      string    `json:"body"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
			return UserNote{}, fmt.Errorf("decode: %w", err)
		}

		return UserNote{
			ID:        note.ID,
			AccountID: note.AccountID,
			Title:     note.Title,
			Body:      note.Body,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		}, nil
	case http.StatusNotFound:
		return UserNote{}, ErrNotFound
	case http.StatusBadRequest:
		b, _ := io.ReadAll(resp.Body)
		return UserNote{}, fmt.Errorf("%w: %s", ErrBadRequest, strings.TrimSpace(string(b)))

	default:
		b, _ := io.ReadAll(resp.Body)
		return UserNote{}, fmt.Errorf("%w: %s", ErrUpstream, strings.TrimSpace(string(b)))
	}
}
