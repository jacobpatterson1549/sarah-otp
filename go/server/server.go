// Package server contains HTTP servers for the website.
package server

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type (
	// Server runs the site.
	Server struct {
		Data        interface{}
		httpsServer *http.Server
		httpServer  *http.Server
		Config
		templateFiles []string
	}

	// Config contains fields which describe the server.
	Config struct {
		// Log is used to log errors and other information
		Log *log.Logger
		// Version is used to bust caches of files from older server versions.
		Version string
		// HTTPPort is the TCP port for server http requests.  All traffic is redirected to the https port.
		HTTPPort int
		// HTTPSPORT is the TCP port for server https requests.
		HTTPSPort int
		// The public HTTPS certificate file.
		TLSCertFile string
		// The private HTTPS key file.
		TLSKeyFile string
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
	case cfg.HTTPPort <= 0:
		return nil, fmt.Errorf("invalid HTTP port: %v", cfg.HTTPPort)
	case cfg.HTTPSPort <= 0:
		return nil, fmt.Errorf("invalid HTTPS port: %v", cfg.HTTPSPort)
	}
	data := map[string]string{
		"Version":         cfg.Version,
		"Name":            "Sarah-OTP",
		"ShortName":       "S-OTP",
		"Description":     "a secure message-passing app",
		"ThemeColor":      "purple",
		"BackgroundColor": "white",
	}
	httpServeMux := new(http.ServeMux)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: httpServeMux,
	}
	httpsServeMux := new(http.ServeMux)
	httpsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPSPort),
		Handler: httpsServeMux,
	}
	templateFiles, err := templateFiles()
	if err != nil {
		return nil, fmt.Errorf("loading template file names: %v", err)
	}
	s := Server{
		Data:          data,
		httpServer:    httpServer,
		httpsServer:   httpsServer,
		Config:        cfg,
		templateFiles: templateFiles,
	}
	httpsServeMux.HandleFunc("/", s.handleHTTPS)
	httpServeMux.HandleFunc("/", s.redirectToHTTPS) // all requests must be HTTPS
	return &s, nil
}

// Run the server asynchronously until it receives a shutdown signal.
// When the HTTP/HTTPS servers stop, errors are logged to the error channel.
func (s Server) Run(ctx context.Context) <-chan error {
	errC := make(chan error, 2)
	go s.runHTTPServer(ctx, errC)
	go s.runHTTPSServer(ctx, errC)
	return errC
}

// Stop asks the server to shutdown and waits for the shutdown to complete.
// An error is returned if the server if the context times out.
func (s Server) Stop(ctx context.Context) error {
	stopDur := 5 * time.Second
	ctx, cancelFunc := context.WithTimeout(ctx, stopDur)
	defer cancelFunc()
	httpsShutdownErr := s.httpsServer.Shutdown(ctx)
	httpShutdownErr := s.httpServer.Shutdown(ctx)
	switch {
	case httpsShutdownErr != nil:
		return httpsShutdownErr
	case httpShutdownErr != nil:
		return httpShutdownErr
	}
	s.Log.Println("server stopped successfully")
	return nil
}

// runHTTPSServer runs the http server, adding the return error to the channel when done.
// The HTTP server is only run if it uses a different port than the HTTPS server.
func (s Server) runHTTPServer(ctx context.Context, errC chan<- error) {
	if !s.httpsOnly() {
		errC <- s.httpServer.ListenAndServe()
	}
}

func (s Server) runHTTPSServer(ctx context.Context, errC chan<- error) {
	switch {
	case s.httpsOnly():
		s.Log.Printf("starting http server at at http://127.0.0.1%v", s.httpsServer.Addr)
		if len(s.TLSCertFile) != 0 || len(s.TLSKeyFile) != 0 {
			s.Log.Printf("Ignoring TLS_CERT_FILE/TLS_KEY_FILE variables since PORT was specified, using automated certificate management.")
		}
		errC <- s.httpsServer.ListenAndServe()
	default:
		s.Log.Printf("starting https server at at https://127.0.0.1%v", s.httpsServer.Addr)
		if _, err := tls.LoadX509KeyPair(s.TLSCertFile, s.TLSKeyFile); err != nil {
			s.Log.Printf("Problem loading tls certificate: %v", err)
			errC <- err
			return
		}
		errC <- s.httpsServer.ListenAndServeTLS(s.TLSCertFile, s.TLSKeyFile)
	}
}

// handleHTTPS handles HTTPS endpoints.
func (s Server) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.TLS == nil && (!s.httpsOnly() || !s.hasSecHeader(r)):
		s.redirectToHTTPS(w, r)
	case r.Method == "GET":
		s.handleHTTPSGet(w, r)
	default:
		s.httpError(w, http.StatusMethodNotAllowed)
	}
}

// redirectToHTTPS redirects the page to HTTPS.
func (s Server) redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if strings.Contains(host, ":") {
		var err error
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			err = fmt.Errorf("could not redirect to https: %w", err)
			s.handleError(w, err)
			return
		}
	}
	if s.HTTPSPort != 443 && !s.httpsOnly() {
		host = host + s.httpsServer.Addr
	}
	httpsURI := "https://" + host + r.URL.Path
	http.Redirect(w, r, httpsURI, http.StatusTemporaryRedirect)
}

// handleHTTPSGet calls handlers for GET endpoints.
func (s Server) handleHTTPSGet(w http.ResponseWriter, r *http.Request) {
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
	case "/", "/serviceWorker.js", "/manifest.json", "/favicon.svg", "/network_check.html":
		s.serveTemplate(w, r, r.URL.Path)
	case "/wasm_exec.js", "/main.wasm":
		http.ServeFile(w, r, "."+r.URL.Path)
	case "/favicon.png", "/robots.txt":
		http.ServeFile(w, r, "resources"+r.URL.Path)
	default:
		s.httpError(w, http.StatusNotFound)
	}
}

// templateFiles gets the list of available resources for templates
func templateFiles() ([]string, error) {
	var filenames []string
	templateFileGlobs := []string{
		"resources/html/*.html",
		"resources/*.js",
		"resources/*.json",
		"resources/*.css",
		"resources/*.svg",
	}
	for _, g := range templateFileGlobs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return nil, err
		}
		filenamesTmp := make([]string, len(filenames)+len(matches))
		copy(filenamesTmp, filenames)
		copy(filenamesTmp[len(filenames):], matches)
		filenames = filenamesTmp
	}
	return filenames, nil
}

// sereveTemplate servers the file from the data-driven template.
func (s Server) serveTemplate(w http.ResponseWriter, r *http.Request, name string) {
	var t *template.Template
	switch name {
	case "/":
		t = template.New("main.html")

	default:
		t = template.New(name[1:])
	}
	if _, err := t.ParseFiles(s.templateFiles...); err != nil {
		err = fmt.Errorf("parsing template %v: %v", name, err)
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

// hasSecHeader returns true if thhe request has any header starting with "Sec-".
func (Server) hasSecHeader(r *http.Request) bool {
	for header := range r.Header {
		if strings.HasPrefix(header, "Sec-") {
			return true
		}
	}
	return false
}

// httpsOnly returns true if the server is only handling HTTPS requests.
func (s Server) httpsOnly() bool {
	return s.HTTPPort == s.HTTPSPort
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
