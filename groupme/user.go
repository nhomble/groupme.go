package groupme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserAPI struct {
	client *Client
}

// GroupMe User Entity
type User struct {
	Id          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	ImageUrl    string `json:"image_url"`
	Name        string `json:"name"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	Email       string `json:"email"`
	Sms         bool   `json:"sms"`
}

// GroupeMe Update User Payload
type UpdateUserCommand struct {
	AvatarUrl *string `json:"avatar_url"`
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	ZipCode   *string `json:"zip_code"`
}

// Get authenticated users information from GroupMe
func (api UserAPI) Get() (*User, error) {
	user := User{}
	url := fmt.Sprintf("%s/users/me?token=%s", BASE, (*api.client.TokenProvider).Get())
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
func (api UserAPI) Update(cmd *UpdateUserCommand) (*User, error) {
	url := fmt.Sprintf("%s/users/update?token=%s", BASE, (*api.client.TokenProvider).Get())
	data, err := json.Marshal(cmd)
	fmt.Println(string(data))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
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
