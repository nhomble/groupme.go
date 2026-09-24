package groupme

import (
	"net/http"
	"time"
)

// UserAPI is the client API responsible for all user functionality.
type UserAPI struct {
	client *Client
}

// User is the GroupMe user data model.
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

// UpdateUserCommand is the request body to update the authenticated user.
type UpdateUserCommand struct {
	AvatarURL *string `json:"avatar_url,omitempty"`
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	ZipCode   *string `json:"zip_code,omitempty"`
}

// Get authenticated users information from GroupMe
func (api UserAPI) Get() (*User, error) {
	user := User{}
	if err := api.client.do(http.MethodGet, "/v3/users/me", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Update users information on GroupMe
func (api UserAPI) Update(cmd UpdateUserCommand) (*User, error) {
	user := &User{}
	if err := api.client.do(http.MethodPost, "/v3/users/update", cmd, user); err != nil {
		return nil, err
	}
	return user, nil
}

// CreatedAtParsed parses the user's created-at epoch time.
func (user User) CreatedAtParsed() time.Time {
	return time.Unix(user.CreatedAt, 0)
}

// UpdatedAtParsed parses the user's updated-at epoch time.
func (user User) UpdatedAtParsed() time.Time {
	return time.Unix(user.UpdatedAt, 0)
}
