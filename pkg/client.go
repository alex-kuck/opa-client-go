package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// NewClient creates a new OPA Client with a baseUrl pointing to an OPA instance
// and using request to perform HTTP requests.
func NewClient(baseUrl string, request RequestDoer) *Client {
	return &Client{
		baseUrl: baseUrl,
		request: request,
	}
}

// RequestDoer is used to perform HTTP requests when communicating with OPA.
type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl string
	request RequestDoer
}

// Query wraps a call to OPA via the client. Everything passed as body will be
// forwarded as input to the OPA rule located at path.
//
// The path = example/allow will access rule "allow" in package "example"
func Query[REQ any, RES any](ctx context.Context, client *Client, path string, body REQ) (res RES, err error) {
	reqUrl, err := url.JoinPath(client.baseUrl, "/v1/data", path)
	if err != nil {
		return res, fmt.Errorf("could not create request url: %w", err)
	}

	req, err := request(ctx, reqUrl, body)
	if err != nil {
		return res, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := client.request.Do(req)
	if err != nil {
		return res, fmt.Errorf("error while performing request: %w", err)
	}

	return response[RES](resp)
}

func request[T any](ctx context.Context, url string, body T) (*http.Request, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(opaRequestBody[T]{Input: body})
	if err != nil {
		return nil, fmt.Errorf("could not encode OPA request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, err
}

type opaRequestBody[T any] struct {
	Input T `json:"input"`
}

func response[T any](response *http.Response) (T, error) {
	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("request to OPA failed with status code %d", response.StatusCode)
		return *new(T), err
	}

	var opaResponse opaResponse[T]
	defer response.Body.Close()
	err := json.NewDecoder(response.Body).Decode(&opaResponse)
	if err != nil {
		return *new(T), fmt.Errorf("could not decode OPA response: %w", err)
	}

	return opaResponse.Result, nil
}

type opaResponse[T any] struct {
	Result T `json:"result"`
}
