// Package server contains HTTP servers for the website.
package server

import (
	"compress/gzip"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type (
	// Server runs the site.
	Server struct {
		Data    any
		server  *http.Server
		tmpl    *template.Template
		Log     *log.Logger
		BuildFS fs.FS
	}

	// Config contains fields which describe the server.
	Config struct {
		// Log is used to log errors and other information
		Log *log.Logger
		// Version is used to bust caches of files from older server versions.
		Version string
		// Port is the TCP port for server http requests.
		Port int
		// ResourcesFS contains the files and templates that are served.
		ResourcesFS fs.FS
		// BuildFS contains binary/build files to be served.
		BuildFS fs.FS
	}

	// wrappedResponseWriter wraps response writing with another writer.
	wrappedResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

// NewServer creates a Server from the Config.
func (cfg Config) NewServer() (*Server, error) {
	switch {
	case cfg.Log == nil:
		return nil, fmt.Errorf("missing logger")
	case cfg.Port <= 0:
		return nil, fmt.Errorf("invalid port: %v", cfg.Port)
	}
	version := strings.TrimSpace(cfg.Version)
	data := map[string]string{
		"Version":         version,
		"Name":            "Sarah-OTP",
		"ShortName":       "S-OTP",
		"Description":     "a secure message-passing app",
		"ThemeColor":      "purple",
		"BackgroundColor": "white",
	}
	serveMux := new(http.ServeMux)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: serveMux,
	}
	t, err := template.ParseFS(cfg.ResourcesFS, "resources/html/*.html", "resources/*.*")
	if err != nil {
		return nil, fmt.Errorf("parsing template filesystem: %v", err)
	}
	s := Server{
		Data:    data,
		server:  server,
		tmpl:    t,
		Log:     cfg.Log,
		BuildFS: cfg.BuildFS,
	}
	serveMux.HandleFunc("/", s.handle)
	return &s, nil
}

// Run the server asynchronously until it receives a shutdown signal.
// When the server stops, errors are logged to the error channel.
func (s Server) Run(ctx context.Context) <-chan error {
	errC := make(chan error, 2)
	go s.runServer(ctx, errC)
	return errC
}

// Stop asks the server to shutdown and waits for the shutdown to complete.
// An error is returned if the server if the context times out.
func (s Server) Stop(ctx context.Context) error {
	stopDur := 5 * time.Second
	ctx, cancelFunc := context.WithTimeout(ctx, stopDur)
	defer cancelFunc()
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	s.Log.Println("server stopped successfully")
	return nil
}

func (s Server) runServer(ctx context.Context, errC chan<- error) {
	s.Log.Printf("starting http server at at http://127.0.0.1%v", s.server.Addr)
	errC <- s.server.ListenAndServe()
}

// handle handles HTTP endpoints.
func (s Server) handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		s.handleGet(w, r)
	default:
		s.httpError(w, http.StatusMethodNotAllowed)
	}
}

// handleGet calls handlers for GET endpoints.
func (s Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w2 := gzip.NewWriter(w)
		defer w2.Close()
		w = wrappedResponseWriter{
			Writer:         w2,
			ResponseWriter: w,
		}
		w.Header().Set("Content-Encoding", "gzip")
	}
	switch r.URL.Path {
	case "/", "/serviceWorker.js", "/manifest.json", "/favicon.svg", "/network_check.html", "/robots.txt":
		s.serveTemplate(w, r, r.URL.Path)
	case "/wasm_exec.js", "/main.wasm":
		http.ServeFileFS(w, r, s.BuildFS, "build"+r.URL.Path)
	case "/favicon.ico":
		// NOOP
	default:
		s.httpError(w, http.StatusNotFound)
	}
}

// serveTemplate servers the file from the data-driven template.
func (s Server) serveTemplate(w http.ResponseWriter, r *http.Request, name string) {
	switch name {
	case "/":
		name = "main.html"
	default:
		name = name[1:]
	}
	t := s.tmpl.Lookup(name)
	if t == nil {
		err := fmt.Errorf("looking up file %v: not found", name)
		s.handleError(w, err)
		return
	}
	s.addMimeType(name, w)
	if err := t.Execute(w, s.Data); err != nil {
		err = fmt.Errorf("rendering template %v: %v", name, err)
		s.handleError(w, err)
		return
	}
}

// httpError writes the error status code.
func (Server) httpError(w http.ResponseWriter, statusCode int) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}

// handleError logs and writes the error as an internal server error (500).
func (s Server) handleError(w http.ResponseWriter, err error) {
	s.Log.Printf("server error: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// addMimeType adds the applicable mime type to the response.
func (Server) addMimeType(fileName string, w http.ResponseWriter) {
	extension := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(extension)
	w.Header().Set("Content-Type", mimeType)
}

// Write delegates the write to the wrapped writer.
func (wrw wrappedResponseWriter) Write(p []byte) (n int, err error) {
	return wrw.Writer.Write(p)
}
