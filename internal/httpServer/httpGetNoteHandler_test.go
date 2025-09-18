package httpserver

import (
	"NotesService/internal/testutil"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPGetNoteHandler(t *testing.T) {
	require.NoError(t, testutil.LoadFixtures(pool, fixturePath))
	sut := HTTPGetNoteHandler(svc)

	type wantReq struct {
		body   string
		header map[string]string
	}
	type wantResp struct {
		code int
		body any
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&testNote)
	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "Wrong mediatype",
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
			name: "Invalid json",
			req: wantReq{
				body:   `{"account_id":101,`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid json"}` + "\n",
			},
		},
		{
			name: "Wrong fields",
			req: wantReq{
				body:   `{"acc_id":101,"ttl":"x"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid json"}` + "\n",
			},
		},
		{
			name: "Invalid id",
			req: wantReq{
				body:   `{"account_id":0,"title":"x","body":"y"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}` + "\n",
			},
		},
		{
			name: "Get existing note",
			req: wantReq{
				body:   `{"account_id":101,"id":1}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusOK,
				body: "noteTest",
			},
		},
		{
			name: "Get non-existing note",
			req: wantReq{
				body:   `{"account_id":101,"id":99999}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"note not found"}` + "\n",
			},
		},
	}
	type reqInfo struct {
		url    string
		method string
	}
	req := reqInfo{
		url:    "/note/get",
		method: http.MethodGet,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(req.method, req.url, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()

			for k, v := range tt.req.header {
				req.Header.Set(k, v)
			}

			sut.ServeHTTP(rr, req)
			assert.Equal(t, tt.want.code, rr.Code, "status code")

			if tt.want.body == "noteTest" {
				var got FixtureNote
				json.Unmarshal(rr.Body.Bytes(), &got)

				assert.Equal(t, testNote, got)
				return
			}

			str, ok := tt.want.body.(string)
			require.True(t, ok, "string assert want body")
			assert.Equal(t, normalizeJSON(t, str), normalizeJSON(t, rr.Body.String()))

		})
	}
}
