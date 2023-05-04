// Created with Strapit
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/logging"
)

var PORT = os.Getenv("PORT")

func main() {
	ctx := context.Background()
	logger, err := NewGoogleCloudLogger(ctx, "holy-diver-297719")
	if err != nil {
		log.Fatal(err)
	}
	flush := logger.SetLogger("dmda-api", logging.RedirectAsJSON(os.Stdout))
	defer flush()

	cache, err := NewDataCache(ctx, "holy-diver-297719")
	if err != nil {
		logger.Log(err, logging.Critical)
	}
	err = cache.Warmup(ctx)
	if err != nil {
		logger.Log(err, logging.Critical)
	}

	signals := make(chan os.Signal, 1)
	defer close(signals)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc("/warmup", WarmupHandler(nil, nil))
	mux.HandleFunc("/data/curated", GetDataHandler(cache, logger))

	// mux.HandleFunc("/warmup", CORSMiddleware(WarmupHandler(nil, nil)))
	// mux.HandleFunc("/data/curated", CORSMiddleware(GetDataHandler(cache, logger)))
	server := &http.Server{
		Addr:        ":" + PORT,
		Handler:     CORSMuxMiddleware(mux),
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error)
	go func() {
		defer close(serveErr)
		logger.Log(EventListening, logging.Info)
		err := server.ListenAndServe()
		if err != nil {
			serveErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			log.Println(err)
			return
		}
	case sig := <-signals:
		logger.Log(EventSignal, logging.Notice)
		switch sig {
		case syscall.SIGHUP:
			// Reserved
		default:
			logger.Log(EventShutdown, logging.Info)
			shutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			err := server.Shutdown(shutCtx)
			if err != nil {
				logger.Log(err, logging.Critical)
			}
		}
	}

}

type Event struct {
	Message string    `json:"message,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

var (
	EventListening *Event = &Event{
		Message: "listening for HTTP(S)",
		Time:    time.Now(),
	}
	EventSignal *Event = &Event{
		Message: "signal received",
		Time:    time.Now(),
	}
	EventShutdown *Event = &Event{
		Message: "initiating graceful shutdown",
		Time:    time.Now(),
	}
	EventWarmup *Event = &Event{
		Message: "warmup request received",
		Time:    time.Now(),
	}
	EventCacheError *Event = &Event{
		Message: "cache error",
		Time:    time.Now(),
	}
	DecodingError *Event = &Event{
		Message: "decoding error",
		Time:    time.Now(),
	}
)
