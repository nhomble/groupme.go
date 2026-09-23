package groupme

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GroupAPI is the client API responsible for all group functionality.
type GroupAPI struct {
	client *Client
}

// Member is a group member.
type Member struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Muted    bool   `json:"muted"`
	ImageURL string `json:"image_url"`
}

// GroupMessages summarizes the message activity of a group.
type GroupMessages struct {
	Count                int            `json:"count"`
	LastMessageID        string         `json:"last_message_id"`
	LastMessageCreatedAt int64          `json:"last_message_created_at"`
	Preview              PreviewMessage `json:"preview"`
}

// Group is the GroupMe group data model.
type Group struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	Description   string        `json:"description"`
	ImageURL      string        `json:"image_url"`
	CreatorUserID string        `json:"creator_user_id"`
	CreatedAt     int64         `json:"created_at"`
	UpdatedAt     int64         `json:"updated_at"`
	Members       []Member      `json:"members"`
	ShareURL      string        `json:"share_url"`
	Messages      GroupMessages `json:"messages"`
}

// GroupQuery controls pagination and field omission when listing groups.
type GroupQuery struct {
	Page    int
	PerPage int
	Omit    []string
}

// CreateGroupCommand is the request body to create a group.
type CreateGroupCommand struct {
	Name     string  `json:"name"`
	Share    bool    `json:"share"`
	ImageURL *string `json:"image_url,omitempty"`
}

// UpdateGroupCommand is the request body to update a group.
type UpdateGroupCommand struct {
	Name       *string `json:"name,omitempty"`
	Share      *bool   `json:"share,omitempty"`
	OfficeMode *bool   `json:"office_mode,omitempty"`
	ImageURL   *string `json:"image_url,omitempty"`
}

var defaultGroupQuery GroupQuery = GroupQuery{
	Page:    1,
	PerPage: 10,
	Omit:    []string{"memberships"},
}

// DefaultGroupQuery returns a fresh copy of the default GroupQuery. Each call
// returns an independent value, including a copy of the Omit slice, so
// callers can safely mutate the returned value without affecting other
// callers.
func DefaultGroupQuery() GroupQuery {
	return GroupQuery{
		Page:    defaultGroupQuery.Page,
		PerPage: defaultGroupQuery.PerPage,
		Omit:    append([]string{}, defaultGroupQuery.Omit...),
	}
}

func (api GroupAPI) searchInternal(endpoint string, q *GroupQuery) ([]Group, error) {
	if q == nil {
		dq := DefaultGroupQuery()
		q = &dq
	}
	if q.PerPage < 0 || q.PerPage > 10 {
		return nil, fmt.Errorf("Invalid number of groups per page=%d", q.PerPage)
	}
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", q.Page))
	values.Set("per_page", fmt.Sprintf("%d", q.PerPage))
	if len(q.Omit) > 0 {
		values.Set("omit", strings.Join(q.Omit, ","))
	}
	path := fmt.Sprintf("/v3%s?%s", endpoint, values.Encode())
	var groups []Group
	if err := api.client.do(http.MethodGet, path, nil, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func forGroup(client *Client, method, path string, in interface{}) (*Group, error) {
	group := Group{}
	if err := client.do(method, path, in, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// Find lists groups the authenticated user is part of.
func (api GroupAPI) Find(q *GroupQuery) ([]Group, error) {
	return api.searchInternal("/groups", q)
}

// FindAll lists all groups the authenticated user is part of, paginating
// through every page.
func (api GroupAPI) FindAll() ([]Group, error) {
	groups := []Group{}
	for i := 1; ; i += 1 {
		q := DefaultGroupQuery()
		q.Page = i
		partial, err := api.Find(&q)
		if err != nil {
			return nil, err
		}
		groups = append(groups, partial...)
		if len(partial) == 0 {
			break
		}
	}
	return groups, nil
}

// FindFormer lists groups the authenticated user was a part of (but can rejoin).
func (api GroupAPI) FindFormer(q *GroupQuery) ([]Group, error) {
	return api.searchInternal("/groups/former", q)
}

// Get group by id
func (api GroupAPI) Get(id string) (*Group, error) {
	return forGroup(api.client, http.MethodGet, fmt.Sprintf("/v3/groups/%s", url.PathEscape(id)), nil)
}

// Create creates a new group.
func (api GroupAPI) Create(cmd CreateGroupCommand) (*Group, error) {
	return forGroup(api.client, http.MethodPost, "/v3/groups", cmd)
}

// Update a group by id
func (api GroupAPI) Update(groupId string, cmd UpdateGroupCommand) (*Group, error) {
	return forGroup(api.client, http.MethodPost, fmt.Sprintf("/v3/groups/%s/update", url.PathEscape(groupId)), cmd)
}

// Delete the group by id
func (api GroupAPI) Delete(groupId string) error {
	return api.client.do(http.MethodPost, fmt.Sprintf("/v3/groups/%s/destroy", url.PathEscape(groupId)), nil, nil)
}

// Join a group for the first time
func (api GroupAPI) Join(groupId string, shareUrl string) (*Group, error) {
	return forGroup(api.client, http.MethodPost, fmt.Sprintf("/v3/groups/%s/join/%s", url.PathEscape(groupId), url.PathEscape(shareUrl)), nil)
}

// ReJoin rejoins a group this user had previously joined.
func (api GroupAPI) ReJoin(groupId string) (*Group, error) {
	cmd := struct {
		ID string `json:"group_id"`
	}{
		ID: groupId,
	}
	return forGroup(api.client, http.MethodPost, "/v3/groups/join", cmd)
}

// CreatedAtParsed parses the group's created-at epoch time.
func (group Group) CreatedAtParsed() time.Time {
	return time.Unix(group.CreatedAt, 0)
}

// UpdatedAtParsed parses the group's updated-at epoch time.
func (group Group) UpdatedAtParsed() time.Time {
	return time.Unix(group.UpdatedAt, 0)
}

// LastMessageCreatedAtParsed parses the last message's created-at epoch time.
func (group GroupMessages) LastMessageCreatedAtParsed() time.Time {
	return time.Unix(group.LastMessageCreatedAt, 0)
}
