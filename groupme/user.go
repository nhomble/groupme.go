package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type UserAPI struct {
	client *Client
}

// GroupMe User Entity
type User struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	ImageURL    string `json:"image_url"`
	Name        string `json:"name"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	Email       string `json:"email"`
	Sms         bool   `json:"sms"`
}

// GroupeMe Update User Payload
type UpdateUserCommand struct {
	AvatarURL *string `json:"avatar_url,omitempty"`
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	ZipCode   *string `json:"zip_code,omitempty"`
}

// Get authenticated users information from GroupMe
func (api UserAPI) Get() (*User, error) {
	user := User{}
	url := api.client.makeURL("/v3/users/me")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update users information on GroupMe
func (api UserAPI) Update(cmd UpdateUserCommand) (*User, error) {
	url := api.client.makeURL("/v3/users/update")
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	user := &User{}
	data, err = api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Parse the time since epoch time from groupme
func (user User) CreatedAtParsed() time.Time {
	return time.Unix(user.CreatedAt, 0)
}

// Parse the time since epoch time from groupme
func (user User) UpdatedAtParsed() time.Time {
	return time.Unix(user.UpdatedAt, 0)
}
