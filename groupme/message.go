package groupme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type MessageAPI struct {
	client *Client
}

type PreviewMessage struct {
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}

type Attachment struct {
}

type Message struct {
	Id          string       `json:"id"`
	SourceGuid  string       `json:"source_guid"`
	CreatedAt   int64        `json:"created_at"`
	UserId      string       `json:"user_id"`
	GroupId     string       `json:"group_id"`
	Name        string       `json:"name"`
	AvatarUrl   string       `json:"avatar_url"`
	Text        string       `json:"text"`
	System      bool         `json:"system"`
	FavoritedBy []string     `json:"favorited_by"`
	Attachments []Attachment `json:"attachments"`
}

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

var DefaultMessageQuery MessageQuery = MessageQuery{
	nil, nil, nil, nil,
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
		lastId = &resp.Messages[len(resp.Messages)-1].Id
	}

	return &MessageIndex{Count: count, Messages: ret}, nil
}

// Build the request URL used by both Query and queryForSearch.
func (api MessageAPI) buildQueryURL(groupId string, q *MessageQuery) (string, error) {
	if q == nil {
		q = &DefaultMessageQuery
	}
	before := ""
	if q.BeforeId != nil {
		before = "&before_id=" + *q.BeforeId
	}
	since := ""
	if q.SinceId != nil {
		since = "&since_id=" + *q.SinceId
	}
	after := ""
	if q.AfterId != nil {
		after = "&after_id=" + *q.AfterId
	}
	limit := fmt.Sprintf("&limit=%d", DEFAULT_MESSAGE_LIMIT)
	if q.Limit != nil {
		if *q.Limit < 0 {
			return "", errors.New(fmt.Sprintf("Provided limit=%d is less than 0!", *q.Limit))
		} else if *q.Limit > 100 {
			return "", errors.New(fmt.Sprintf("Provided limit=%d is greater than 100!", *q.Limit))
		}
		limit = fmt.Sprintf("&limit=%d", *q.Limit)
	}
	return api.client.makeURL(fmt.Sprintf("/v3/groups/%s/messages?%s%s%s%s", groupId, before, since, after, limit)), nil
}

// Get messages in the group
func (api MessageAPI) Query(groupId string, q *MessageQuery) (*MessageIndex, error) {
	url, err := api.buildQueryURL(groupId, q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
func (api MessageAPI) Send(groupId string, cmd *SendMessageCommand) (*Message, error) {
	url := api.client.makeURL(fmt.Sprintf("/v3/groups/%s/messages", groupId))
	data, err := json.Marshal(struct {
		Message SendMessageCommand `json:"message"`
	}{Message: *cmd})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
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
