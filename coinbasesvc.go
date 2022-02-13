package main

import (
	"encoding/json"
	"io"
	"net/http"
)

const DefaultURL = "https://api.coinbase.com/v2/prices/spot"

type CoinbaseSvc struct {
	URL        string
	httpClient *http.Client
}

type Response struct {
	Data   *ResponseData    `json:"data,omitempty"`
	Errors *[]ResponseError `json:"errors,omitempty"`
}

type ResponseData struct {
	Base     string `json:"base"`
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type ResponseError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ServiceError struct {
	StatusCode int
	InnerError ResponseError
}

func NewCoinbaseSvc(url string) CoinbaseSvc {
	return CoinbaseSvc{
		URL:        url,
		httpClient: &http.Client{},
	}
}

func (client CoinbaseSvc) GetSpotPrice(currency string) (*Response, *ServiceError) {
	r, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			InnerError: ResponseError{"internal_error", err.Error()},
		}
	}
	q := r.URL.Query()
	q.Add("currency", currency)
	r.URL.RawQuery = q.Encode()

	resp, err := client.httpClient.Do(r)
	if err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			InnerError: ResponseError{"internal_error", err.Error()},
		}
	}
	defer resp.Body.Close()
	jsonResp := &Response{}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			InnerError: ResponseError{"internal_error", err.Error()},
		}
	}
	if err = json.Unmarshal(b, jsonResp); err != nil {
		return nil, &ServiceError{
			StatusCode: http.StatusInternalServerError,
			InnerError: ResponseError{"internal_error", err.Error()},
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &ServiceError{
			StatusCode: resp.StatusCode,
			InnerError: (*jsonResp.Errors)[0],
		}
	}

	return jsonResp, nil
}
