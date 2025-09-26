package httpserver

import (
	"net/http"
	"net/http/httptest"
	"notes/internal/gateway/testutil"
	userservice "notes/internal/gateway/userService"
	"notes/internal/pkg/middlewares"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_createNoteHandler(t *testing.T) {
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
				body:   `{"account_id":101,"title":"Hank","body":"Bears"}`,
			},
			want: wantResp{
				code: http.StatusCreated,
				body: `{"id":1}`,
			},
		},
		{
			name: "02_Wrong_mediatype",
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
			name: "03_Invalid_json",
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
			name: "04_Wrong_fields",
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
			name: "05_Invalid_id",
			req: wantReq{
				body:   `{"account_id":0,"title":"x","body":"y"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}`,
			},
		},
		{
			name: "06_Empty_title",
			req: wantReq{
				body:   `{"account_id":0,"title":"x","body":"y"}`,
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
		createNoteHandler(svc))

	method := http.MethodPost
	hndURL := "/note/create"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("Create", mock.Anything, mock.AnythingOfType("int"), mock.Anything, mock.Anything).
				Return(1, nil).Once()
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
