package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type LambdaEvent struct {
	Currency string `json:"currency"`
}

type LambdaResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func writeError(w http.ResponseWriter, statusCode int, id, msg string) {
	bytes, err := json.Marshal(
		Response{
			Errors: &[]ResponseError{{ID: id, Message: msg}},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(bytes)
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "Invalid request")
}

// spotPriceHandler proxies requests through to coinbase spot price api
func createSpotPriceHandler(svc CoinbaseSvc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resp, err := svc.GetSpotPrice(vars["currency"])
		if err != nil {
			writeError(w, err.StatusCode, err.InnerError.ID, err.InnerError.Message)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_ = json.NewEncoder(w).Encode(*resp)
	}
}

// health returns 200 whilst webserver is running
func healthHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func lambdaHandler(svc CoinbaseSvc) func(ctx context.Context, event LambdaEvent) (*LambdaResponse, error) {
	return func(ctx context.Context, event LambdaEvent) (*LambdaResponse, error) {
		resp, err := svc.GetSpotPrice(event.Currency)
		if err != nil {
			b, e := json.Marshal(
				Response{
					Errors: &[]ResponseError{err.InnerError},
				},
			)
			if e != nil {
				return nil, e
			}
			return &LambdaResponse{
				StatusCode: err.StatusCode,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: string(b),
			}, nil
		}
		b, e := json.Marshal(resp)
		if e != nil {
			return nil, e
		}
		return &LambdaResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(b),
		}, nil
	}
}
