package main

import (
	"context"
	"github.com/foae/dimago/clients/github"
	"github.com/foae/dimago/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	cctx, ccancel := context.WithCancel(ctx)
	defer ccancel()
	_ = cctx

	/*
		Check with ENV
	*/
	addr := mustGetEnv("HTTP_LISTEN_ADDR")
	env := mustGetEnv("ENV")
	waitClose := 1
	if env == "prod" {
		waitClose = 5
		// @TODO: reduce logging in production to `Warn` and up.
	}

	/*
		Build up the config
	*/
	githubChan := make(chan bool)
	githubClient := github.NewClient(githubChan)
	hdlr := handler.NewHandler(handler.Config{
		GithubClient: githubClient,
	})

	/*
		Start the HTTP server
	*/
	go func() {
		log.Printf("HTTP Server is up and running on %v", addr)
		if err := http.ListenAndServe(addr, hdlr); err != http.ErrServerClosed {
			log.Fatalf("server: could not serve: %v", err)
		}
	}()

	/*
		--------------
		All systems go
		--------------
	*/

	// Allow cleanup before closing:
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP)
	<-sig

	log.Printf("Shutting down in (%v) second(s)...", waitClose)
	ccancel()
	time.Sleep(time.Second * time.Duration(waitClose))
	log.Println("BYE!")
}

func mustGetEnv(value string) string {
	env := os.Getenv(value)
	if env == "" {
		log.Fatalf("Environment variable `%v` must be set.", value)
	}

	return env
}
