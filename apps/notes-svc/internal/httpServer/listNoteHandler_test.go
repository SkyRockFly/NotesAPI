package httpserver

import (
	"net/http"
	"net/http/httptest"
	"notes/apps/notes-svc/internal/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPListNoteHandler(t *testing.T) {
	require.NoError(t, testutil.LoadFixtures(pool, fixturePath))
	sut := HTTPListNoteHandler(svc)

	type wantReq struct {
		body   string
		header map[string]string
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
			name: "01_Wrong_mediatype",
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
			name: "02_Invalid_json",
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
			name: "03_Wrong_fields",
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
			name: "04_Invalid_id",
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
			name: "05_Get_existing_notes",
			req: wantReq{
				body:   `{"account_id":101}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusOK,
				body: `[{"id":1,"account_id":101,"title":"Первый след когтей","body":
"Бруня оставил заметку прямо на дверце — когтями. Данные для теста GET.",
"created_at":"2025-08-12T08:00:00Z","updated_at":"2025-08-12T08:15:00Z"},{"id":2,"account_id":101,"title":"Мёд и сталь","body"
:"Заметка про то, как он обмакнул лапу в мёд и пошёл ковыряться в коде.","created_at":"2025-08-12T09:30:00Z","updated_at":"2025-08-12T09:45:00Z"}]`,
			},
		},
		{
			name: "06_Get_non-existing_notes",
			req: wantReq{
				body:   `{"account_id":1010101}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"note not found"}`,
			},
		},
	}

	url := "/note/create"
	method := http.MethodPost

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(method, url, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()

			for k, v := range tt.req.header {
				req.Header.Set(k, v)
			}

			sut.ServeHTTP(rr, req)
			assert.Equal(t, tt.want.code, rr.Code, "status code")
			assert.Equal(t, normalizeJSON(t, tt.want.body), normalizeJSON(t, rr.Body.String()))

		})
	}
}
