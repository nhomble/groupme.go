package groupme

import (
	"fmt"
	"io/ioutil"
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

// Common request function
func (c *Client) getResponse(req *http.Request) ([]byte, error) {
	token, err := c.TokenProvider.Get()
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !successful(resp.StatusCode) {
		return nil, newAPIError(req.Method, req.URL.String(), resp.StatusCode, data)
	}

	return data, nil
}

// getResponseWithStatus behaves like getResponse but also returns the raw
// HTTP status code, and treats any status listed in allowedStatuses as a
// non-error response (its body, if any, is returned as-is).
func (c *Client) getResponseWithStatus(req *http.Request, allowedStatuses ...int) ([]byte, int, error) {
	token, err := c.TokenProvider.Get()
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if successful(resp.StatusCode) {
		return data, resp.StatusCode, nil
	}

	for _, s := range allowedStatuses {
		if resp.StatusCode == s {
			return data, resp.StatusCode, nil
		}
	}

	return nil, resp.StatusCode, newAPIError(req.Method, req.URL.String(), resp.StatusCode, data)
}

// Execute request with no expected return value
func (c *Client) execute(req *http.Request) error {
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

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !successful(resp.StatusCode) {
		return newAPIError(req.Method, req.URL.String(), resp.StatusCode, data)
	}
	return nil
}

// Format urls with override considerations
func (c *Client) makeURL(path string) string {
	base := fmt.Sprintf("https://%s", c.host)
	return fmt.Sprintf("%s%s", base, path)
}
