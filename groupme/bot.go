package groupme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type BotAPI struct {
	client *Client
}

type BotMessageCommand struct {
	BotId      string  `json:"bot_id"`
	Message    string  `json:"text"`
	PictureUrl *string `json:"picture_url"`
}

// Send message from bot
func (api BotAPI) Send(cmd BotMessageCommand) error {
	url := fmt.Sprintf("%s/bots/post", BASE)
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
