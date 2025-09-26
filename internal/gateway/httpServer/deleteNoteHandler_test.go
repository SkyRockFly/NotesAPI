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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_deleteNoteHandler(t *testing.T) {
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
				body:   `{"id":2,"account_id":101}`,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"deleted":true}`,
			},
		},
		{
			name: "02_Non-existing note",
			req: wantReq{
				header: map[string]string{"Content-Type": "application/json"},
				body:   `{"id":3,"account_id":101}`,
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
			name: "06_Invalid_id",
			req: wantReq{
				body:   `{"id":0,"account_id":101}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}`,
			},
		},
		{
			name: "07_Invalid_AccountID",
			req: wantReq{
				body:   `{"id":2,"account_id":0}`,
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
		deleteNoteHandler(svc))

	method := http.MethodDelete
	hndURL := "/note/delete"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("Delete", mock.Anything, 2, 101).
				Return(true, nil).Once()
			repo.On("Delete", mock.Anything, 3, 101).
				Return(false, usernoterepo.ErrNotFound).Once()
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
