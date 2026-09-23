package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const UserAgent = "groupme.go/api"

// defaultHTTPTimeout bounds outbound requests so a stalled connection
// cannot block a caller forever.
const defaultHTTPTimeout = 30 * time.Second

// GroupMe SDK client
type Client struct {
	httpClient    *http.Client
	host          string
	TokenProvider TokenProvider
	Users         *UserAPI
	Groups        *GroupAPI
	Messages      *MessageAPI
	Bots          *BotAPI
}

// Returns a new instance to a groupme client
//	token provider
func NewClient(provider TokenProvider) (*Client, error) {
	if _, err := provider.Get(); err != nil {
		return nil, fmt.Errorf("invalid token provider: %w", err)
	}

	httpClient := &http.Client{Timeout: defaultHTTPTimeout}
	c := &Client{httpClient: httpClient}
	c.TokenProvider = provider

	// apis
	c.Users = &UserAPI{client: c}
	c.Groups = &GroupAPI{client: c}
	c.Messages = &MessageAPI{client: c}
	c.Bots = &BotAPI{client: c}
	c.host = "api.groupme.com"

	return c, nil
}

// Set your own http.Client and fluently return the Client
func (c *Client) SetHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// SetHost overrides the default host and fluently return the Client
func (c *Client) SetHost(host string) *Client {
	c.host = host
	return c
}

func successful(code int) bool {
	return code < 300 && code >= 200
}

// do builds and executes an HTTP request against the GroupMe API, and is the
// single consolidated entry point used by every API method.
//
// When in is non-nil, it is marshaled to JSON and sent as the request body;
// GET/DELETE-style calls with no body should pass nil. When the response is
// successful (2xx, or a status listed in allowedExtraStatuses) and out is
// non-nil, the response body is unraveled into out. Any status not in the
// 2xx range and not listed in allowedExtraStatuses results in a typed
// *APIError.
func (c *Client) do(method, path string, in, out interface{}, allowedExtraStatuses ...int) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, c.makeURL(path), body)
	if err != nil {
		return err
	}

	token, err := c.TokenProvider.Get()
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !successful(resp.StatusCode) {
		for _, s := range allowedExtraStatuses {
			if resp.StatusCode == s {
				return nil
			}
		}
		return newAPIError(req.Method, req.URL.String(), resp.StatusCode, data)
	}

	if out != nil {
		return unravel(data, out)
	}
	return nil
}

// Format urls with override considerations
func (c *Client) makeURL(path string) string {
	base := fmt.Sprintf("https://%s", c.host)
	return fmt.Sprintf("%s%s", base, path)
}
