package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSpotPrice(t *testing.T) {
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
			name:     "invalid json",
			response: []byte(`{"invalid":"json"`),
			err:      &ServiceError{StatusCode: 500, InnerError: ResponseError{ID: "internal_error", Message: "unexpected end of JSON input"}},
		},
	}

	// Create mock upstream server
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := req.URL.Query().Get("currency")
			_, _ = w.Write(tt[c].response)
		}),
	)
	defer s.Close()

	client := NewCoinbaseSvc(s.URL)

	for currency, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r, err := client.GetSpotPrice(currency)
			if err != nil {
				assert.Equal(t, tt[currency].err, err)
			} else {
				resp := &Response{}
				_ = json.Unmarshal(tt[currency].response, resp)
				assert.Equal(t, r, resp)
			}
		})
	}
}
