package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacobpatterson1549/sarah-otp/server"
)

// main creates and runs the server.
func main() {
	m := newMainFlags(os.Args, os.LookupEnv)
	log := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()
	server, err := createServer(ctx, m, log)
	if err != nil {
		log.Fatal(err)
	}
	runServer(ctx, *server, log)
}

// create server greates the server from a configuration.
func createServer(ctx context.Context, m mainFlags, log *log.Logger) (*server.Server, error) {
	version, err := version(m.versionFile)
	if err != nil {
		return nil, fmt.Errorf("reading versionFile: %v", err)
	}
	cfg := server.Config{
		Log:         log,
		Version:     version,
		HTTPPort:    m.httpPort,
		HTTPSPort:   m.httpsPort,
		TLSCertFile: m.tlsCertFile,
		TLSKeyFile:  m.tlsKeyFile,
	}
	server, err := cfg.NewServer()
	if err != nil {
		return nil, fmt.Errorf("creating server: %v", err)
	}
	return server, nil
}

// runServer runs the server until it is interrupted or terminated.
func runServer(ctx context.Context, server server.Server, log *log.Logger) {
	done := make(chan os.Signal, 2)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	errC := server.Run(ctx)
	select { // BLOCKING
	case err := <-errC:
		switch {
		case err == http.ErrServerClosed:
			log.Printf("server shutdown triggered")
		default:
			log.Printf("server stopped unexpectedly: %v", err)
		}
	case signal := <-done:
		log.Printf("handled %v", signal)
	}
	if err := server.Stop(ctx); err != nil {
		log.Printf("stopping server: %v", err)
	}
}

// version reads the first word of the versionFile to use as the version.
func version(versionFileName string) (string, error) {
	versionFile, err := os.Open(versionFileName)
	if err != nil {
		return "", fmt.Errorf("trying to open version file: %v", err)
	}
	scanner := bufio.NewScanner(versionFile)
	scanner.Split(bufio.ScanWords)
	if !scanner.Scan() {
		return "", fmt.Errorf("no words in version file")
	}
	return scanner.Text(), nil
}
