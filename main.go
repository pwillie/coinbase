package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type apiHandler struct{}

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router(),
	}

	if err := gracefulShutdown(srv, 10*time.Second); err != nil {
		log.Println(err)
	}
}

func router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", http.HandlerFunc(health))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", apiHandler{})
	return mux
}

func (apiHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if len(req.URL.Path) != 4 {
		http.NotFound(w, req)
		return
	}
	r, err := http.NewRequest("GET", "https://api.coinbase.com/v2/prices/spot", nil)
	if err != nil {
		http.Error(w, "Something bad happened", http.StatusInternalServerError)
		return
	}

	q := req.URL.Query()
	q.Add("currency", req.URL.Path[1:])
	r.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		http.Error(w, "Something bad happened", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

func health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
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
