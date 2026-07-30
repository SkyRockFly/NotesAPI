package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomValidators(t *testing.T) {
	v, err := initValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		value   string
		tag     string
		wantErr bool
	}{
		{
			name:    "#01_VALID_DSN",
			value:   "postgres://postgres:postgres@localhost:5432/notes",
			tag:     "dsn",
			wantErr: false,
		},
		{
			name:    "#02_EMPTY_DSN",
			value:   "",
			tag:     "dsn",
			wantErr: true,
		},
		{
			name:    "#03_DSN_WITHOUT_SCHEME",
			value:   "localhost:5432/notes",
			tag:     "dsn",
			wantErr: true,
		},
		{
			name:    "#04_DSN_WITHOUT_HOST",
			value:   "postgres:///notes",
			tag:     "dsn",
			wantErr: true,
		},
		{
			name:    "#05_VALID_LOG_LEVEL",
			value:   "info",
			tag:     "loglevel",
			wantErr: false,
		},
		{
			name:    "#06_LOG_LEVEL_WITH_SPACES",
			value:   "  debug  ",
			tag:     "loglevel",
			wantErr: false,
		},
		{
			name:    "#07_INVALID_LOG_LEVEL",
			value:   "lmao",
			tag:     "loglevel",
			wantErr: true,
		},
		{
			name:    "#08_INVALID_LOG_LEVEL",
			value:   "lmao",
			tag:     "loglevel",
			wantErr: true,
		},
		{
			name:    "#09_VALID_FORMAT_LEVEL",
			value:   "[%s]",
			tag:     "formatlevel",
			wantErr: false,
		},
		{
			name:    "#10_INVALID_FORMAT_LEVEL",
			value:   "%d",
			tag:     "formatlevel",
			wantErr: true,
		},
		{
			name:    "#11_VALID_TIMESTAMP",
			value:   "2006-01-02 15:04:05",
			tag:     "timestamp",
			wantErr: false,
		},
		{
			name:    "#12_INVALID_TIMESTAMP",
			value:   "02-01-2026",
			tag:     "timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := struct {
				Value string `validate:"required"`
			}{
				Value: tt.value,
			}

			err := v.Var(req.Value, tt.tag)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetSecretKey(t *testing.T) {
	tests := []struct {
		name    string
		content string
		noFile  bool
		wantErr bool
	}{
		{
			name:    "#01_OK",
			content: "12345678901234567890123456789012",
		},
		{
			name:    "#02_TRIM_SPACES",
			content: "  12345678901234567890123456789012\n",
		},
		{
			name:    "#03_WEAK_KEY",
			content: "short",
			wantErr: true,
		},
		{
			name:    "#04_FILE_NOT_FOUND",
			noFile:  true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "jwt-secret")

			if !tt.noFile {
				require.NoError(
					t,
					os.WriteFile(path, []byte(tt.content), 0600),
				)
			}

			auth := &AuthConfig{
				JWTSecretFile: path,
			}

			err := GetSecretKey(auth)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, auth.JWTSecret, 32)
		})
	}
}
