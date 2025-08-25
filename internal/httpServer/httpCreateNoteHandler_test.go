package httpserver_test

import (
	noteservice "NotesService/internal/noteService"
	"net/http"
	"reflect"
	"testing"
)

func TestHTTPCreateNoteHandler(t *testing.T) {
	type args struct {
		service *noteservice.Service
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPCreateNoteHandler(tt.args.service); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPCreateNoteHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
