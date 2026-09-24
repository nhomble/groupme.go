package groupme

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const UserAgent = "groupme.go/api"

// defaultHTTPTimeout bounds outbound requests so a stalled connection
// cannot block a caller forever.
const defaultHTTPTimeout = 30 * time.Second

// maxResponseBodySize bounds how much of a response body is read into
// memory, so a misbehaving or malicious host (e.g. via SetHost) cannot
// exhaust memory with an oversized response.
const maxResponseBodySize = 10 << 20 // 10 MiB

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
	if provider == nil {
		return nil, errors.New("token provider must not be nil")
	}
	if _, err := provider.Get(); err != nil {
		return nil, fmt.Errorf("invalid token provider: %w", err)
	}

	httpClient := &http.Client{Timeout: defaultHTTPTimeout, CheckRedirect: stripSensitiveHeadersOnRedirect}
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

// Set your own http.Client and fluently return the Client. A nil client is
// ignored, since it would otherwise panic on the next request. Note that a
// user-supplied client bypasses the X-Access-Token redirect protection
// NewClient sets up on its own default client.
func (c *Client) SetHTTPClient(client *http.Client) *Client {
	if client == nil {
		return c
	}
	c.httpClient = client
	return c
}

// SetHost overrides the default host and fluently return the Client.
//
// Not safe for concurrent use with in-flight requests: configure the client
// before sharing it across goroutines.
func (c *Client) SetHost(host string) *Client {
	c.host = host
	return c
}

// stripSensitiveHeadersOnRedirect removes the access token header before a
// redirect is followed to a different host, so it is never sent to a host
// other than the one the caller configured.
func stripSensitiveHeadersOnRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	if req.URL.Host != via[0].URL.Host {
		req.Header.Del("X-Access-Token")
	}
	return nil
}

func successful(code int) bool {
	return code < 300 && code >= 200
}

// doRequest sets the common headers, executes req, and reads its body
// (capped at maxResponseBodySize). It is the shared core of getResponse,
// getResponseWithStatus, and execute.
func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	token, err := c.TokenProvider.Get()
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", UserAgent)
	if req.Method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Access-Token", token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

// Common request function
func (c *Client) getResponse(req *http.Request) ([]byte, error) {
	data, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	if !successful(status) {
		return nil, newAPIError(req.Method, req.URL.String(), status, data)
	}
	return data, nil
}

// getResponseWithStatus behaves like getResponse but also returns the raw
// HTTP status code, and treats any status listed in allowedStatuses as a
// non-error response (its body, if any, is returned as-is).
func (c *Client) getResponseWithStatus(req *http.Request, allowedStatuses ...int) ([]byte, int, error) {
	data, status, err := c.doRequest(req)
	if err != nil {
		return nil, status, err
	}
	if successful(status) {
		return data, status, nil
	}
	for _, s := range allowedStatuses {
		if status == s {
			return data, status, nil
		}
	}
	return nil, status, newAPIError(req.Method, req.URL.String(), status, data)
}

// Execute request with no expected return value
func (c *Client) execute(req *http.Request) error {
	data, status, err := c.doRequest(req)
	if err != nil {
		return err
	}
	if !successful(status) {
		return newAPIError(req.Method, req.URL.String(), status, data)
	}
	return nil
}

// Format urls with override considerations
func (c *Client) makeURL(path string) string {
	base := fmt.Sprintf("https://%s", c.host)
	return fmt.Sprintf("%s%s", base, path)
}
