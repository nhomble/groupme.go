package groupme

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Client api responsible for all bot functionality
type BotAPI struct {
	client *Client
}

// Request command to send messages as a bot
type BotMessageCommand struct {
	BotID      string  `json:"bot_id"`
	Message    string  `json:"text"`
	PictureURL *string `json:"picture_url"`
}

// Send message from bot
func (api BotAPI) Send(cmd BotMessageCommand) error {
	url := api.client.makeUrl("/v3/bots/post")
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = api.client.getResponse(req)
	if err != nil {
		return err
	}
	return nil
}
