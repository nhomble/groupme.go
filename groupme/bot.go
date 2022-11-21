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

// Request body to create a bot
type CreateBotCommand struct {
	Name        string  `json:"name"`
	GroupId     string  `json:"group_id"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	CallbackURL *string `json:"callback_url,omitempty"`
}

type createBotCommandRequest struct {
	Bot CreateBotCommand `json:"bot"`
}

// Bot data model in GroupMe
type BotDefitionWithGroupId struct {
	Name          string  `json:"name"`
	GroupId       string  `json:"group_id"`
	AvatarUrl     *string `json:"avatar_url"`
	CallbackUrl   *string `json:"callback_url"`
	Notifications bool    `json:"dm_notification"`
	BotId         string  `json:"bot_id"`
}

type bot struct {
	Bot BotDefitionWithGroupId `json:"bot"`
}

type deleteBotCommand struct {
	BotId string `json:"bot_id"`
}

// Send message from bot
func (api BotAPI) Send(cmd BotMessageCommand) error {
	url := api.client.makeURL("/v3/bots/post")
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

func (api BotAPI) Create(cmd CreateBotCommand) (*BotDefitionWithGroupId, error) {
	url := api.client.makeURL("/v3/bots")
	envelope := createBotCommandRequest{
		Bot: cmd,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	bot := bot{}
	data, err = api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &bot)
	if err != nil {
		return nil, err
	}
	return &bot.Bot, nil
}

func (api BotAPI) List() ([]BotDefitionWithGroupId, error) {
	url := api.client.makeURL("/v3/bots")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	var bots []BotDefitionWithGroupId
	err = unravel(&data, &bots)
	if err != nil {
		return nil, err
	}
	return bots, nil
}

func (api BotAPI) Delete(botId string) error {
	url := api.client.makeURL("/v3/bots/destroy")
	data, err := json.Marshal(deleteBotCommand{
		BotId: botId,
	})
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
