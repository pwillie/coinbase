package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "localhost:8080/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	rec := httptest.NewRecorder()
	defaultHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status not found; got %v", res.Status)
	}
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "localhost:8080/health", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	rec := httptest.NewRecorder()
	healthHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}
}

func TestSpotPriceHandler(t *testing.T) {
	tt := map[string]struct {
		name     string
		response []byte
		err      *ServiceError
	}{
		"aud": {
			name:     "happy path",
			response: []byte(`{"data":{"base":"BTC","currency":"AUD","amount":"59465.39029629"}}`),
		},
		"aaa": {
			name:     "invalid currency",
			response: []byte(`{"errors":[{"id":"not_found","message":"Invalid currency"}]}`),
			err:      &ServiceError{StatusCode: 404, InnerError: ResponseError{ID: "not_found", Message: "Invalid currency"}},
		},
	}

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := req.URL.Query().Get("currency")
			if tt[c].err != nil {
				w.WriteHeader(tt[c].err.StatusCode)
			}
			_, _ = w.Write(tt[c].response)
		}),
	)
	defer s.Close()

	client := NewCoinbaseSvc(s.URL)

	handler := createSpotPriceHandler(client)

	for currency, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://localhost:8080/"+currency, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rec := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/{currency:[A-Za-z]{3}}", handler)
			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if tc.err != nil {
				assert.Equal(t, tc.err.StatusCode, res.StatusCode)
			}

			assert.Equal(t, tc.response, bytes.TrimSpace(b))

			// bytes.TrimSpace(b)
			// if msg := string(bytes.TrimSpace(b)); msg != tc.response {
			// 	t.Errorf("expected message %q; got %q", tc.response, msg)
			// }

			// d, err := strconv.Atoi(string(bytes.TrimSpace(b)))
			// if err != nil {
			// 	t.Fatalf("expected an integer; got %s", b)
			// }
			// if d != tc.double {
			// 	t.Fatalf("expected double to be %v; got %v", tc.double, d)
			// }
		})
	}
}
