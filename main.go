package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	svc := NewCoinbaseSvc(DefaultURL)
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		r := mux.NewRouter()
		r.HandleFunc("/health", http.HandlerFunc(healthHandler))
		r.Handle("/metrics", promhttp.Handler())
		r.HandleFunc("/{currency:[A-Za-z]{3}}", createSpotPriceHandler(svc))
		r.PathPrefix("/").HandlerFunc(defaultHandler)
		r.Use(loggingMiddleware)

		srv := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}

		if err := gracefulShutdown(srv, 10*time.Second); err != nil {
			log.Println(err)
		}
	} else {
		lambda.Start(lambdaHandler(svc))
	}
}

// gracefulShutdown stops the given HTTP server on
// receiving a stop signal and waits for the active connections
// to be closed for {timeout} period of time.
func gracefulShutdown(srv *http.Server, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		ctx := context.Background()
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		done <- srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-done
}
