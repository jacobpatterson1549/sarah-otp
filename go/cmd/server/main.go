// Package server runs the website.
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacobpatterson1549/sarah-otp/go/server"
)

// resourcesFS is the embedded resources directory.
//
//go:embed resources
var resourcesFS embed.FS

// buildFS is the embedded build directory.
//
//go:embed build
var buildFS embed.FS

// version is the server version
//
//go:embed build/version
var version string

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

// create server creates the server from a configuration.
func createServer(ctx context.Context, m mainFlags, log *log.Logger) (*server.Server, error) {
	cfg := server.Config{
		Log:         log,
		Version:     version,
		Port:        m.Port,
		ResourcesFS: resourcesFS,
		BuildFS:     buildFS,
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
