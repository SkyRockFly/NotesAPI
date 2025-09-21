package usernoterepo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpRepo struct {
	base   string
	client *http.Client
}

func NewHTTPRepo(base string) *httpRepo {
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

	reqURL, err := url.JoinPath(s.base, "note/get")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := s.sendReq(ctx, req, http.MethodPost, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	note, err := decodeUserNote(resp)
	if err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func (s *httpRepo) Create(ctx context.Context, accountID int, title, body string) (UserNote, error) {
	req := struct {
		AccountID int    `json:"account_id"`
		Title     string `json:"title"`
		Body      string `json:"body"`
	}{accountID, title, body}

	reqURL, err := url.JoinPath(s.base, "note/create")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := s.sendReq(ctx, req, http.MethodPost, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	note, err := decodeUserNote(resp)
	if err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func (s *httpRepo) Update(ctx context.Context, id, accountID int, title, body string) (UserNote, error) {
	req := struct {
		ID        int    `json:"id"`
		AccountID int    `json:"account_id"`
		Title     string `json:"title"`
		Body      string `json:"body"`
	}{id, accountID, title, body}

	reqURL, err := url.JoinPath(s.base, "note/update")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := s.sendReq(ctx, req, http.MethodPatch, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	note, err := decodeUserNote(resp)
	if err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func (s *httpRepo) Delete(ctx context.Context, id, accountID int) (UserNote, error) {
	req := struct {
		ID        int `json:"id"`
		AccountID int `json:"account_id"`
	}{id, accountID}

	reqURL, err := url.JoinPath(s.base, "note/delete")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := s.sendReq(ctx, req, http.MethodDelete, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	note, err := decodeUserNote(resp)
	if err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func (s *httpRepo) List(ctx context.Context, accountID int) (UserNote, error) {
	req := struct {
		AccountID int `json:"account_id"`
	}{accountID}

	reqURL, err := url.JoinPath(s.base, "notes/get")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := s.sendReq(ctx, req, http.MethodPost, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	note, err := decodeUserNote(resp)
	if err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func mapRespErrors(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		b, _ := io.ReadAll(r.Body)
		return fmt.Errorf("%w: %s", ErrNotFound, strings.TrimSpace(string(b)))
	case http.StatusBadRequest:
		b, _ := io.ReadAll(r.Body)
		return fmt.Errorf("%w: %s", ErrBadRequest, strings.TrimSpace(string(b)))
	default:
		b, _ := io.ReadAll(r.Body)
		return fmt.Errorf("%w: %s", ErrUpstream, strings.TrimSpace(string(b)))
	}
}

func decodeUserNote(r *http.Response) (UserNote, error) {
	var note struct {
		ID        int       `json:"id"`
		AccountID int       `json:"account_id"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
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
}

func (s *httpRepo) sendReq(ctx context.Context, body any, method, url string) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method,
		url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}
