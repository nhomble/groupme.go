package groupme

import (
	"errors"
	"io/ioutil"
	"net/http"
)

const AGENT = "groupme.go/api"

// GroupMe SDK client
type Client struct {
	httpClient    *http.Client
	TokenProvider *TokenProvider
	Users         *UserAPI
	Groups        *GroupAPI
	Messages      *MessageAPI
	Bots          *BotAPI
}

// Returns a new instance to a groupme client
//	token provider
func NewClient(provider TokenProvider) (*Client, error) {
	httpClient := http.DefaultClient
	c := &Client{httpClient: httpClient}
	c.TokenProvider = &provider

	// apis
	c.Users = &UserAPI{client: c}
	c.Groups = &GroupAPI{client: c}
	c.Messages = &MessageAPI{client: c}
	c.Bots = &BotAPI{client: c}

	return c, nil
}

// Set your own http.Client and fluently return the Client
func (c *Client) SetHttpClient(client http.Client) *Client {
	c.httpClient = &client
	return c
}

func successful(code int) bool {
	return code < 300 && code >= 200
}

// Common request function
func (c *Client) getResponse(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", AGENT)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", (*c.TokenProvider).Get())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if !successful(resp.StatusCode) {
		if resp.StatusCode == 400 {
			return nil, errors.New(parseError(&data))
		} else {
			return nil, errors.New("Failed to make " + req.Method + " request to url=" + req.URL.String() + " status=" + resp.Status)
		}
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Execute request with no expected return value
func (c *Client) execute(req *http.Request) error {
	req.Header.Set("User-Agent", AGENT)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", (*c.TokenProvider).Get())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if !successful(resp.StatusCode) {
		if resp.StatusCode == 400 {
			return errors.New(parseError(&data))
		} else {
			return errors.New("Failed to make " + req.Method + " request to url=" + req.URL.String() + " status=" + resp.Status)
		}
	}
	return nil
}
