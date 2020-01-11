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
	Count                int            `json:"count"`
	LastMessageId        string         `json:"last_message_id"`
	LastMessageCreatedAt int64          `json:"last_message_created_at"`
	Preview              PreviewMessage `json:"preview"`
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

type UpdateGroupCommand struct {
	Name       *string `json:"name"`
	Share      bool    `json:"share"`
	OfficeMode bool    `json:"office_mode"`
	ImageUrl   *string `json:"image_url"`
}

var DefaultGroupQuery GroupQuery = GroupQuery{
	Page:    1,
	PerPage: 10,
	Omit:    []string{"memberships"},
}

func (api GroupAPI) searchInternal(endpoint string, q *GroupQuery) ([]Group, error) {
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
	url := fmt.Sprintf("%s%s?page=%d&per_page=%d%s", BASE, endpoint, q.Page, q.PerPage, omit)
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

func forGroup(client *Client, req *http.Request) (*Group, error) {
	group := Group{}
	data, err := client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// List groups the authenticated user is part of
func (api GroupAPI) Find(q *GroupQuery) ([]Group, error) {
	return api.searchInternal("/groups", q)
}

func (api GroupAPI) FindAll() ([]Group, error) {
	do := true
	groups := []Group{}
	for i := 1; do; i += 1 {
		q := GroupQuery{
			Page:    i,
			PerPage: 10,
			Omit:    []string{"memberships"},
		}
		partial, err := api.Find(&q)
		if err != nil {
			return nil, err
		}
		if len(partial) == 0 {
			do = false
		}

		groups = append(groups, partial...)
	}
	return groups, nil
}

// List groups the authenticated user was a part of (but can rejoin)
func (api GroupAPI) FindFormer(q *GroupQuery) ([]Group, error) {
	return api.searchInternal("/groups/former", q)
}

// Get group by id
func (api GroupAPI) Get(id string) (*Group, error) {
	url := fmt.Sprintf("%s/groups/%s", BASE, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return forGroup(api.client, req)
}

func (api GroupAPI) Create(cmd *CreateGroupCommand) (*Group, error) {
	url := fmt.Sprintf("%s/groups", BASE)
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return forGroup(api.client, req)
}

// Update a group by id
func (api GroupAPI) Update(groupId string, cmd *UpdateGroupCommand) (*Group, error) {
	url := fmt.Sprintf("%s/groups/%s/update", BASE, groupId)
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return forGroup(api.client, req)
}

// Delete the group by id
func (api GroupAPI) Delete(groupId string) error {
	url := fmt.Sprintf("%s/groups/%s/destroy", BASE, groupId)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	return api.client.execute(req)
}

// Join a group for the first time
func (api GroupAPI) Join(groupId string, shareUrl string) (*Group, error) {
	url := fmt.Sprintf("%s/groups/%s/join/%s", BASE, groupId, shareUrl)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	return forGroup(api.client, req)
}

// Rejoin a group this user had previously joined
func (api GroupAPI) ReJoin(groupId string) (*Group, error) {
	url := fmt.Sprintf("%s/groups/join", BASE)
	data, err := json.Marshal(struct {
		Id string `json:"group_id"`
	}{
		Id: groupId,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return forGroup(api.client, req)
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
