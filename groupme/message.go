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
	total := -1
	seen := 0
	count := 0
	var ret []Message
	var lastId *string
	lastId = nil
	for total == -1 || (seen < total && !search.StopCriteria(count, total, seen)) {
		resp, err := api.Query(groupId, &MessageQuery{
			BeforeId: lastId,
		})

		if err != nil {
			return nil, err
		}

		if total == -1 {
			total = resp.Count
		}

		for _, message := range resp.Messages {
			if search.Criteria(message) {
				count += 1
				ret = append(ret, message)
			}
			seen += 1
		}
		lastId = &resp.Messages[len(resp.Messages)-1].Id
	}

	return &MessageIndex{Count: count, Messages: ret}, nil
}

// Get messages in the group
func (api MessageAPI) Query(groupId string, q *MessageQuery) (*MessageIndex, error) {
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
	limit := "&limit=20"
	if q.Limit != nil {
		if *q.Limit < 0 {
			return nil, errors.New(fmt.Sprintf("Provided limit=%d is less than 0!", *q.Limit))
		} else if *q.Limit > 10 {
			return nil, errors.New(fmt.Sprintf("Provided limit=%d is greater than 10!", *q.Limit))
		}
		limit = fmt.Sprintf("&limit=%d", q.Limit)
	}
	url := api.client.makeUrl(fmt.Sprintf("/v3/groups/%s/messages?%s%s%s%s", groupId, before, since, after, limit))
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

// Send a message to the group
func (api MessageAPI) Send(groupId string, cmd *SendMessageCommand) (*Message, error) {
	url := api.client.makeUrl(fmt.Sprintf("/v3/groups/%s/messages", groupId))
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
