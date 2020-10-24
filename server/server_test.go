package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	newServerTests := []struct {
		config Config
		wantOk bool
	}{
		{}, // missing Logger
		{ // The HTTPS port must be defined
			config: Config{
				Log:      log,
				HTTPPort: 80,
			},
		},
		{ // The HTTP port must be defined even if only HTTPS is used.
			config: Config{
				Log:       log,
				HTTPSPort: 443,
			},
		},
		{ // This configuration should be used for running locally.
			config: Config{
				Log:       log,
				HTTPPort:  80,
				HTTPSPort: 443,
			},
			wantOk: true,
		},
		{ // This is the configuration Heroku should use.
			config: Config{
				Log:       log,
				HTTPPort:  8001,
				HTTPSPort: 8001,
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
		case server.Log == nil, server.httpServer == nil, server.httpsServer == nil:
			t.Errorf("test %v: wanted log, httpServer, httpsServer to not be nil", i)
		}
	}
}

func TestHasSecHeader(t *testing.T) {
	hasSecHeaderTests := map[string]bool{
		"Accept":          false,
		"DNT":             false,
		"":                false,
		"inSec-t":         false,
		"Sec-Fetch-Mode:": true,
	}
	var s Server
	for header, want := range hasSecHeaderTests {
		r := http.Request{
			Header: map[string][]string{
				header: nil,
			},
		}
		got := s.hasSecHeader(&r)
		if want != got {
			t.Errorf("wanted hasSecHeader = %v when header = %v", want, header)
		}
	}
}

func TestHTTPSOnly(t *testing.T) {
	httpsOnlyTests := []struct {
		httpPort  int
		httpsPort int
		want      bool
	}{
		{
			want: true,
		},
		{
			httpPort:  80,
			httpsPort: 443,
		},
		{
			httpPort:  8001,
			httpsPort: 8001,
			want:      true,
		},
	}
	for i, test := range httpsOnlyTests {
		s := Server{
			Config: Config{
				HTTPPort:  test.httpPort,
				HTTPSPort: test.httpsPort,
			},
		}
		got := s.httpsOnly()
		if test.want != got {
			t.Errorf("test %v: wanted httpsOnly = %v when httpPort= %v and httpsPort = %v", i, test.want, test.httpPort, test.httpsPort)
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
		Config: Config{
			Log: log.New(&buf, "", log.LstdFlags),
		},
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
		"favicon.png":   "image/png",
		"favicon.svg":   "image/svg+xml",
		"manifest.json": "application/json",
		"init.js":       "application/javascript",
		"main.wasm":     "application/wasm",
		"/":             "", // no mime type
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
