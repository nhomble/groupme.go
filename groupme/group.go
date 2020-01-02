package groupme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GroupAPI struct {
	client *Client
}

type Member struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Muted    bool   `json:"muted"`
	ImageUrl string `json:"image_url"`
}

type GroupMessages struct {
	Count                int     `json:"count"`
	LastMessageId        string  `json:"last_message_id"`
	LastMessageCreatedAt int64   `json:"last_message_created_at"`
	Preview              Message `json:"preview"`
}

type Group struct {
	Id            string        `json:"id"`
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	Description   string        `json:"description"`
	ImageUrl      string        `json:"image_url"`
	CreatorUserId string        `json:"creator_user_id"`
	CreatedAt     int64         `json:"created_at"`
	UpdatedAt     int64         `json:"updated_at"`
	Members       []Member      `json:"members"`
	ShareUrl      string        `json:"share_url"`
	Messages      GroupMessages `json:"messages"`
}

type GroupQuery struct {
	Page    int
	PerPage int
	Omit    []string
}

type CreateGroupCommand struct {
	Name     string  `json:"name"`
	Share    bool    `json:"share"`
	ImageUrl *string `json:"image_url"`
}

var DefaultGroupQuery GroupQuery = GroupQuery{
	Page:    1,
	PerPage: 10,
	Omit:    []string{"memberships"},
}

func (api GroupAPI) getInternal(endpoint string, q *GroupQuery) ([]Group, error) {
	if q == nil {
		q = &DefaultGroupQuery
	}
	if q.PerPage < 0 || q.PerPage > 10 {
		return nil, errors.New(fmt.Sprintf("Invalid number of groups per page=%d", q.PerPage))
	}
	omit := ""
	if len(q.Omit) > 0 {
		omit = "&omit=" + strings.Join(q.Omit, ",")
	}
	url := fmt.Sprintf("%s%s?token=%s&page=%d&per_page=%d%s", BASE, endpoint, api.client.token, q.Page, q.PerPage, omit)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)

	if err != nil {
		return nil, err
	}
	var groups []Group
	err = unravel(&data, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// List groups the authenticated user is part of
func (api GroupAPI) Find(q *GroupQuery) ([]Group, error) {
	return api.getInternal("/groups", q)
}

// List groups the authenticated user was a part of (but can rejoin)
func (api GroupAPI) FindFormer(q *GroupQuery) ([]Group, error) {
	return api.getInternal("/groups/former", q)
}

// Get group by id
func (api GroupAPI) Get(id string) (*Group, error) {
	url := fmt.Sprintf("%s/groups/%s?token=%s", BASE, id, api.client.token)
	group := Group{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (api GroupAPI) Create(cmd *CreateGroupCommand) (*Group, error) {
	url := fmt.Sprintf("%s/groups?token=%s", BASE, api.client.token)
	data, err := json.Marshal(cmd)
	fmt.Println(string(data))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	group := Group{}
	data, err = api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil

}

// Parse the time since epoch time from groupme
func (group Group) CreatedAtParsed() time.Time {
	return time.Unix(group.CreatedAt, 0)
}

// Parse the time since epoch time from groupme
func (group Group) UpdatedAtParsed() time.Time {
	return time.Unix(group.UpdatedAt, 0)
}

// Parse the time since epoch time from groupme
func (group GroupMessages) LastMessageCreatedAtParsed() time.Time {
	return time.Unix(group.LastMessageCreatedAt, 0)
}
