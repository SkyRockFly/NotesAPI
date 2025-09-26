package httpserver

import (
	"net/http"
	"net/http/httptest"
	"notes/internal/gateway/testutil"
	usernoterepo "notes/internal/gateway/userNoteRepo"
	userservice "notes/internal/gateway/userService"
	"notes/internal/pkg/middlewares"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_listNoteHandler(t *testing.T) {
	type wantReq struct {
		header map[string]string
		body   string
	}
	type wantResp struct {
		code int
		body string
	}

	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "01_OK",
			req: wantReq{
				header: map[string]string{"Content-Type": "application/json"},
				body:   `{"account_id":101}`,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `[{"id":1,"account_id":101,"title":"kek","body":"cheburek","created_at":"2025-09-12T10:00:00Z",
"updated_at":"2025-09-12T10:00:00Z"},{"id":2,"account_id":101,"title":"giga",
"body":"chad","created_at":"2025-09-12T11:00:00Z","updated_at":"2025-09-12T11:00:00Z"}]`,
			},
		},
		{
			name: "02_Non-existing_accountID",
			req: wantReq{
				header: map[string]string{"Content-Type": "application/json"},
				body:   `{"account_id":202}`,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"note not found"}`,
			},
		},
		{
			name: "03_Wrong_mediatype",
			req: wantReq{
				body:   `{"account_id":101,"title":"Honey","body":"Bears"}`,
				header: map[string]string{"Content-Type": "text/plain"},
			},
			want: wantResp{
				code: http.StatusUnsupportedMediaType,
				body: "unsupported media type\n",
			},
		},
		{
			name: "04_Invalid_json",
			req: wantReq{
				body:   `{"account_id":101,`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "05_Wrong_fields",
			req: wantReq{
				body:   `{"acc_id":101,"ttl":"x"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "06_Invalid_AccountID",
			req: wantReq{
				body:   `{"account_id":0}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}`,
			},
		},
	}

	repo := new(mockNoteRepo)
	svc := userservice.NewService(repo)

	sut := middlewares.DemandJSONHeaders(
		listNoteHandler(svc))

	method := http.MethodPost
	hndURL := "/notes/get"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("List", mock.Anything, 101).
				Return([]usernoterepo.UserNote{
					{
						ID:        1,
						AccountID: 101,
						Title:     "kek",
						Body:      "cheburek",
						CreatedAt: time.Date(2025, 9, 12, 10, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2025, 9, 12, 10, 0, 0, 0, time.UTC),
					},
					{
						ID:        2,
						AccountID: 101,
						Title:     "giga",
						Body:      "chad",
						CreatedAt: time.Date(2025, 9, 12, 11, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2025, 9, 12, 11, 0, 0, 0, time.UTC),
					},
				}, nil).Once()
			repo.On("List", mock.Anything, 202).
				Return([]usernoterepo.UserNote{}, usernoterepo.ErrNotFound).Once()
			req := httptest.NewRequest(method, hndURL, strings.NewReader(tt.req.body))

			for k, v := range tt.req.header {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			sut.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body),
				testutil.NormalizeJSON(t, rr.Body.String()))
		})
	}
}
