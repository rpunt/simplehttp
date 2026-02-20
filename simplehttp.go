package simplehttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

type HTTPClient struct {
	BaseURL string
	Headers map[string]string
	Data    map[string]string
	Params  map[string]string
	Client  *http.Client
}

type HTTPResponse struct {
	Body    string
	Code    int
	Headers map[string][]string
}

func New(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Headers: make(map[string]string),
		Data:    make(map[string]string),
		Params:  make(map[string]string),
		Client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (client *HTTPClient) SetTimeout(timeout time.Duration) {
	client.Client.Timeout = timeout
}

func sendRequest(client *HTTPClient, path string, method string) (HTTPResponse, error) {
	if client.Client == nil {
		return HTTPResponse{}, fmt.Errorf("simplehttp: %s %s: http client is nil", method, path)
	}

	// create the request body, as appropriate
	var requestData []byte
	if len(client.Data) > 0 {
		var err error
		requestData, err = json.Marshal(client.Data)
		if err != nil {
			return HTTPResponse{}, fmt.Errorf("simplehttp: %s %s: marshaling request data: %w", method, path, err)
		}
	}

	// construct the request
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", client.BaseURL, path), bytes.NewBuffer(requestData))
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("simplehttp: %s %s: creating request: %w", method, path, err)
	}
	for k, v := range client.Headers {
		req.Header.Set(k, v)
	}

	// add query params, if any
	q := req.URL.Query()
	for k, v := range client.Params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// do :allthethings:
	response, err := client.Client.Do(req) //nolint:gosec // URL is caller-provided by design
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("simplehttp: %s %s: %w", method, path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("simplehttp: %s %s: reading response body: %w", method, path, err)
	}

	responseHeaders := make(map[string][]string)
	for k, v := range response.Header {
		responseHeaders[k] = v
	}

	resp := HTTPResponse{
		Body:    string(body),
		Code:    response.StatusCode,
		Headers: responseHeaders,
	}
	return resp, nil
}

func (client *HTTPClient) Get(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodGet)
}

func (client *HTTPClient) Post(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodPost)
}

func (client *HTTPClient) Patch(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodPatch)
}

func (client *HTTPClient) Put(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodPut)
}

func (client *HTTPClient) Delete(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodDelete)
}

func (client *HTTPClient) Head(path string) (HTTPResponse, error) {
	return sendRequest(client, path, http.MethodHead)
}
