package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type MessageAPI struct {
	client *Client
}

type PreviewMessage struct {
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}

// Attachment represents a single attachment on a GroupMe message. GroupMe
// attachments are polymorphic: the Type field determines which of the
// remaining fields are populated (e.g. "image" uses URL, "location" uses
// Lat/Lng/Name, "mentions" uses UserIDs/Loci, "emoji" uses Placeholder/
// Charmap). Fields not applicable to a given Type are left zero-valued.
type Attachment struct {
	Type        string   `json:"type"`
	URL         string   `json:"url,omitempty"`
	Lat         string   `json:"lat,omitempty"`
	Lng         string   `json:"lng,omitempty"`
	Name        string   `json:"name,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
	Loci        [][]int  `json:"loci,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Charmap     [][]int  `json:"charmap,omitempty"`
}

type Message struct {
	ID          string       `json:"id"`
	SourceGuid  string       `json:"source_guid"`
	CreatedAt   int64        `json:"created_at"`
	UserID      string       `json:"user_id"`
	GroupID     string       `json:"group_id"`
	Name        string       `json:"name"`
	AvatarURL   string       `json:"avatar_url"`
	Text        string       `json:"text"`
	System      bool         `json:"system"`
	FavoritedBy []string     `json:"favorited_by"`
	Attachments []Attachment `json:"attachments"`
}

// MessageIndex holds a page of messages. Count's meaning depends on how it
// was obtained: from Query, it is the group's total message count; from
// Search, it is instead the number of messages that matched the criteria.
type MessageIndex struct {
	Count    int       `json:"count"`
	Messages []Message `json:"messages"`
}

type MessageQuery struct {
	BeforeId *string
	SinceId  *string
	AfterId  *string
	Limit    *int
}

type MessageSearch struct {
	Limit        *int
	Criteria     func(message Message) bool
	StopCriteria func(count int, total int, seen int) bool
}

var defaultMessageQuery MessageQuery = MessageQuery{
	nil, nil, nil, nil,
}

// DefaultMessageQuery returns a fresh copy of the default MessageQuery. Each
// call returns an independent value so callers can safely mutate the
// returned value without affecting other callers.
func DefaultMessageQuery() MessageQuery {
	return MessageQuery{
		BeforeId: defaultMessageQuery.BeforeId,
		SinceId:  defaultMessageQuery.SinceId,
		AfterId:  defaultMessageQuery.AfterId,
		Limit:    defaultMessageQuery.Limit,
	}
}

type SendMessageCommand struct {
	SourceGuid  string       `json:"source_guid"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func (api MessageAPI) Search(groupId string, search MessageSearch) (*MessageIndex, error) {
	criteria := search.Criteria
	if criteria == nil {
		criteria = func(message Message) bool { return true }
	}
	stopCriteria := search.StopCriteria
	if stopCriteria == nil {
		stopCriteria = func(count int, total int, seen int) bool { return false }
	}

	total := -1
	seen := 0
	count := 0
	var ret []Message
	var lastId *string
	lastId = nil
	for total == -1 || (seen < total && !stopCriteria(count, total, seen)) {
		if search.Limit != nil && count >= *search.Limit {
			break
		}

		resp, status, err := api.queryForSearch(groupId, &MessageQuery{
			BeforeId: lastId,
		})

		if err != nil {
			return nil, err
		}

		if status == http.StatusNotModified || len(resp.Messages) == 0 {
			// No more messages left to paginate through.
			break
		}

		if total == -1 {
			total = resp.Count
		}

		for _, message := range resp.Messages {
			if search.Limit != nil && count >= *search.Limit {
				break
			}
			if criteria(message) {
				count += 1
				ret = append(ret, message)
			}
			seen += 1
		}
		lastId = &resp.Messages[len(resp.Messages)-1].ID
	}

	return &MessageIndex{Count: count, Messages: ret}, nil
}

// Build the request URL used by both Query and queryForSearch.
func (api MessageAPI) buildQueryURL(groupId string, q *MessageQuery) (string, error) {
	if err := validID(groupId); err != nil {
		return "", err
	}
	if q == nil {
		dq := DefaultMessageQuery()
		q = &dq
	}
	values := url.Values{}
	if q.BeforeId != nil {
		values.Set("before_id", *q.BeforeId)
	}
	if q.SinceId != nil {
		values.Set("since_id", *q.SinceId)
	}
	if q.AfterId != nil {
		values.Set("after_id", *q.AfterId)
	}
	limit := DefaultMessageLimit
	if q.Limit != nil {
		if *q.Limit < 0 {
			return "", fmt.Errorf("Provided limit=%d is less than 0!", *q.Limit)
		} else if *q.Limit > 100 {
			return "", fmt.Errorf("Provided limit=%d is greater than 100!", *q.Limit)
		}
		limit = *q.Limit
	}
	values.Set("limit", fmt.Sprintf("%d", limit))
	return api.client.makeURL(fmt.Sprintf("/v3/groups/%s/messages?%s", url.PathEscape(groupId), values.Encode())), nil
}

// Get messages in the group
func (api MessageAPI) Query(groupId string, q *MessageQuery) (*MessageIndex, error) {
	url, err := api.buildQueryURL(groupId, q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	messages := MessageIndex{}
	err = unravel(&data, &messages)
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

// queryForSearch is like Query, but used only by the Search pagination loop:
// it treats HTTP 304 Not Modified (returned by GroupMe when there are no
// older messages left to paginate) as an empty, non-error result rather than
// a hard failure, so Search can stop cleanly and return what it has
// collected so far.
func (api MessageAPI) queryForSearch(groupId string, q *MessageQuery) (*MessageIndex, int, error) {
	url, err := api.buildQueryURL(groupId, q)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	data, status, err := api.client.getResponseWithStatus(req, http.StatusNotModified)
	if err != nil {
		return nil, status, err
	}
	if status == http.StatusNotModified {
		return &MessageIndex{}, status, nil
	}
	messages := MessageIndex{}
	err = unravel(&data, &messages)
	if err != nil {
		return nil, status, err
	}
	return &messages, status, nil
}

// Send a message to the group
func (api MessageAPI) Send(groupId string, cmd SendMessageCommand) (*Message, error) {
	if err := validID(groupId); err != nil {
		return nil, err
	}
	reqURL := api.client.makeURL(fmt.Sprintf("/v3/groups/%s/messages", url.PathEscape(groupId)))
	data, err := json.Marshal(struct {
		Message SendMessageCommand `json:"message"`
	}{Message: cmd})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	data, err = api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	result := struct {
		Message Message `json:"message"`
	}{}
	err = unravel(&data, &result)
	if err != nil {
		return nil, err
	}
	return &result.Message, nil
}

// Parse the time since epoch time from groupme
func (message Message) CreatedAtParsed() time.Time {
	return time.Unix(message.CreatedAt, 0)
}
