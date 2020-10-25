package instagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// defaultBaseURL is a default URL of Instagram API.
	defaultBaseURL = "https://graph.instagram.com"
)

// ConfigOption helps to configure the Client.
type ConfigOption func(*Client)

// WithBaseURL sets Client's Instagram API domain which is useful when testing.
func WithBaseURL(baseURL string) ConfigOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets Client's underlying http.Client.
func WithHTTPClient(httpClient *http.Client) ConfigOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// Client manages communication with the Instagram API.
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// NewClient returns a Client which can be configured with config options.
func NewClient(accessToken string, options ...ConfigOption) *Client {
	c := Client{
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		httpClient:  http.DefaultClient,
	}
	for _, opt := range options {
		opt(&c)
	}
	return &c
}

// NewRequest creates http.Request to access Instagram API.
// API path must not start or end with a slash.
// Query string parameters are optional.
// If specified, the value pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(ctx context.Context, method, path string, queryParams url.Values, bodyParams interface{}) (*http.Request, error) {
	var urlStr string
	if queryParams != nil {
		urlStr = fmt.Sprintf("%s/%s?%s", c.baseURL, path, queryParams.Encode())
	} else {
		urlStr = fmt.Sprintf("%s/%s", c.baseURL, path)
	}

	var (
		b   []byte
		err error
	)
	if bodyParams != nil {
		if b, err = json.Marshal(bodyParams); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, urlStr, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		bearer := fmt.Sprintf("Bearer %s", c.accessToken)
		req.Header.Set("Authorization", bearer)
	}

	return req, nil
}

// Do uses Client's http.Client to execute the http.Request and unmarshals the http.Response into v.
// It also handles unmarshaling errors returned by the server.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	e := Error{
		HTTPStatusCode: resp.StatusCode,
		Body:           string(body),
	}

	if resp.StatusCode == http.StatusOK {
		if err = json.Unmarshal(body, v); err != nil {
			e.Inner = err
			return resp, e
		}
		return resp, nil
	}

	errResp := struct {
		Error `json:"error"`
	}{e}
	if err = json.Unmarshal(body, &errResp); err != nil {
		errResp.Error.Inner = err
	}
	return resp, errResp.Error
}
