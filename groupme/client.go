package groupme

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// GroupMe SDK client
type Client struct {
	httpClient *http.Client
	token      string
	Users      *UserAPI
	Groups     *GroupAPI
}

// Returns a new instance to a groupme client
// 	token 		- groupme api token
//	httpClient 	- compose the http client
func NewClient(token string, httpClient *http.Client) (*Client, error) {
	if len(token) == 0 {
		return nil, errors.New("must provide a groupme token")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{httpClient: httpClient}
	c.token = token

	// apis
	c.Users = &UserAPI{client: c}
	c.Groups = &GroupAPI{client: c}

	return c, nil
}

func successful(code int) bool {
	return code < 300 && code >= 200
}

// Common request function
func (c *Client) getResponse(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", "groupme.go/api")
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
			return nil, errors.New("Failed to make http request status=" + resp.Status)
		}
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}
