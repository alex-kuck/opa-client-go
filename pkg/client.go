package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func NewClient(baseUrl string, request RequestDoer) *Client {
	return &Client{
		baseUrl: baseUrl,
		request: request,
	}
}

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl string
	request RequestDoer
}

const restApiPrefix = "/v1/data"

func Query[REQ any, RES any](ctx context.Context, client *Client, path string, body REQ) (res RES, err error) {
	reqUrl, err := url.JoinPath(client.baseUrl, restApiPrefix, path)
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
