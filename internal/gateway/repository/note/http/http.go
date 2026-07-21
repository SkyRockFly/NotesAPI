package notehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	noterepo "notes/internal/gateway/repository/note"
	"notes/internal/pkg/apperror"
	"time"
)

type HTTPRepo struct {
	baseURL string
	client  *http.Client
}

type SvcHTTPCfg struct {
	Host string
	Port int
}

type CreateResp struct {
	ID int `json:"id"`
}

func NewHTTPRepo(cfg SvcHTTPCfg) *HTTPRepo {
	u := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}

	return &HTTPRepo{
		baseURL: u.String(),
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *HTTPRepo) Get(ctx context.Context, req noterepo.GetReq) (noterepo.Note, error) {
	reqURL, err := url.JoinPath(s.baseURL, "note/get")
	if err != nil {
		return noterepo.Note{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPost, reqURL)
	if err != nil {
		return noterepo.Note{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return noterepo.Note{}, fmt.Errorf("response: %w", err)
	}

	var note noterepo.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return noterepo.Note{}, fmt.Errorf("decode: %w", err)
	}

	return note, nil
}

func (s *HTTPRepo) Create(ctx context.Context, req noterepo.CreateReq) (int, error) {
	reqURL, err := url.JoinPath(s.baseURL, "note/create")
	if err != nil {
		return 0, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPost, reqURL)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return 0, fmt.Errorf("response: %w", err)
	}

	var repoResp CreateResp
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}

	return repoResp.ID, nil
}

func (s *HTTPRepo) Update(ctx context.Context, req noterepo.UpdateReq) error {
	reqURL, err := url.JoinPath(s.baseURL, "note/update")
	if err != nil {
		return fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPatch, reqURL)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return fmt.Errorf("response: %w", err)
	}

	return nil
}

func (s *HTTPRepo) Delete(ctx context.Context, req noterepo.DeleteReq) error {
	reqURL, err := url.JoinPath(s.baseURL, "note/delete")
	if err != nil {
		return fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodDelete, reqURL)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return fmt.Errorf("response: %w", err)
	}

	return nil
}

func (s *HTTPRepo) List(ctx context.Context, req noterepo.ListReq) (noterepo.ListResp, error) {
	reqURL, err := url.JoinPath(s.baseURL, "notes/get")
	if err != nil {
		return noterepo.ListResp{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPost, reqURL)
	if err != nil {
		return noterepo.ListResp{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return noterepo.ListResp{}, fmt.Errorf("response: %w", err)
	}

	var dto noterepo.ListResp
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return noterepo.ListResp{}, fmt.Errorf("decode: %w", err)
	}

	return dto, nil
}

func mapRespErrors(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return apperror.ErrNotFound
	case http.StatusBadRequest:
		return apperror.ErrBadRequest
	default:
		return apperror.ErrBackend
	}
}

func sendReq(ctx context.Context, client *http.Client, body any, method, url string) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}
