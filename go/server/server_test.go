package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestNewServer(t *testing.T) {
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	newServerTests := []struct {
		config Config
		wantOk bool
	}{
		{}, // missing Logger
		{
			config: Config{
				Log:  log,
				Port: 8001,
				ResourcesFS: &fstest.MapFS{
					"resources/html/main.html": {},
					"resources/main.css":       {},
				},
			},
			wantOk: true,
		},
	}
	for i, test := range newServerTests {
		server, err := test.config.NewServer()
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("test %v: unwanted error: %v", i, err)
		case server.Log == nil, server.server == nil:
			t.Errorf("test %v: wanted log, server to not be nil", i)
		}
	}
}

func TestHTTPError(t *testing.T) {
	w := httptest.NewRecorder()
	var s Server
	want := 400
	s.httpError(w, want)
	got := w.Code
	switch {
	case want != got:
		t.Errorf("wanted error message to contain %v, got %v", want, got)
	case w.Body.Len() <= 1: // ends in \n character
		t.Errorf("wanted status code info for error (%v) in body", want)
	}
}

func TestHandleError(t *testing.T) {
	var buf bytes.Buffer
	w := httptest.NewRecorder()
	err := fmt.Errorf("mock error")
	s := Server{
		Log: log.New(&buf, "", log.LstdFlags),
	}
	want := 500
	s.handleError(w, err)
	got := w.Code
	switch {
	case want != got:
		t.Errorf("wanted error message to contain %v, got %v", want, got)
	case !strings.Contains(w.Body.String(), err.Error()):
		t.Errorf("wanted message in body (%v), but got %v", err.Error(), w.Body.String())
	case !strings.Contains(buf.String(), err.Error()):
		t.Errorf("wanted message in log (%v), but got %v", err.Error(), buf.String())
	}
}

func TestAddMimeType(t *testing.T) {
	addMimeTypeTests := map[string]string{
		"favicon.svg": "image/svg+xml",
		"main.wasm":   "application/wasm",
		"/":           "", // no mime type
	}
	for fileName, want := range addMimeTypeTests {
		var s Server
		w := httptest.NewRecorder()
		s.addMimeType(fileName, w)
		got := w.Header().Get("Content-Type")
		if want != got {
			t.Errorf("when filename = %v, wanted mimeType %v, got %v", fileName, want, got)
		}
	}
}
