package usernoterepo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// dto requests

type createDTO struct {
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type getDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type deleteDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type updateDTO struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type listDTO struct {
	AccountID int `json:"account_id"`
}

// dto responses
type createResp struct {
	ID int `json:"id"`
}

type noteResp struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type deleteResp struct {
	Deleted bool `json:"deleted"`
}

type updateResp struct {
	Updated bool `json:"updated"`
}

///

type httpRepo struct {
	baseURL string
	client  *http.Client
}

type SvcHTTPCfg struct {
	Host string
	Port int
}

func NewHTTPRepo(cfg SvcHTTPCfg) *httpRepo {
	u := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}

	return &httpRepo{
		baseURL: u.String(),
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

func (s *httpRepo) Get(ctx context.Context, id int, accountID int) (UserNote, error) {
	req := getDTO{
		ID:        id,
		AccountID: accountID,
	}

	reqURL, err := url.JoinPath(s.baseURL, "note/get")
	if err != nil {
		return UserNote{}, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPost, reqURL)
	if err != nil {
		return UserNote{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return UserNote{}, fmt.Errorf("response: %w", err)
	}

	var note noteResp
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return UserNote{}, fmt.Errorf("decode: %w", err)
	}

	userNote := toUserNote(note)

	return userNote, nil
}

func (s *httpRepo) Create(ctx context.Context, accountID int, title, body string) (int, error) {
	req := createDTO{
		AccountID: accountID,
		Title:     title,
		Body:      body,
	}

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

	var answ createResp
	if err := json.NewDecoder(resp.Body).Decode(&answ); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}

	return answ.ID, nil
}

func (s *httpRepo) Update(ctx context.Context, id, accountID int, title, body string) (bool, error) {
	req := updateDTO{
		ID:        id,
		AccountID: accountID,
		Title:     title,
		Body:      body,
	}

	reqURL, err := url.JoinPath(s.baseURL, "note/update")
	if err != nil {
		return false, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPatch, reqURL)
	if err != nil {
		return false, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return false, fmt.Errorf("response: %w", err)
	}

	var answ updateResp
	if err := json.NewDecoder(resp.Body).Decode(&answ); err != nil {
		return false, fmt.Errorf("decode: %w", err)
	}

	return answ.Updated, nil
}

func (s *httpRepo) Delete(ctx context.Context, id, accountID int) (bool, error) {
	req := deleteDTO{
		ID:        id,
		AccountID: accountID,
	}

	reqURL, err := url.JoinPath(s.baseURL, "note/delete")
	if err != nil {
		return false, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodDelete, reqURL)
	if err != nil {
		return false, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return false, fmt.Errorf("response: %w", err)
	}

	var answ deleteResp
	if err := json.NewDecoder(resp.Body).Decode(&answ); err != nil {
		return false, fmt.Errorf("decode: %w", err)
	}

	return answ.Deleted, nil
}

func (s *httpRepo) List(ctx context.Context, accountID int) ([]UserNote, error) {
	req := listDTO{
		AccountID: accountID,
	}

	reqURL, err := url.JoinPath(s.baseURL, "notes/get")
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	resp, err := sendReq(ctx, s.client, req, http.MethodPost, reqURL)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := mapRespErrors(resp); err != nil {
		return nil, fmt.Errorf("response: %w", err)
	}

	var notes []noteResp
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	userNotes := make([]UserNote, 0, len(notes))
	var userNote UserNote
	for _, dtoNote := range notes {
		userNote = toUserNote(dtoNote)
		userNotes = append(userNotes, userNote)
	}

	return userNotes, nil
}

func mapRespErrors(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusCreated:
		fallthrough
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		return ErrBadRequest
	default:
		return ErrUpstream
	}
}

func toUserNote(dto noteResp) UserNote {
	//lint:ignore S1016 держим границу между контрактом и репо + явный маппинг
	return UserNote{
		ID:        dto.ID,
		AccountID: dto.AccountID,
		Title:     dto.Title,
		Body:      dto.Body,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func sendReq(ctx context.Context, client *http.Client, body any, method, url string) (*http.Response, error) {
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

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}
