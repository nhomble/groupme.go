package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrBotNotFound is returned when no bot matches the given ID
var ErrBotNotFound = errors.New("bot not found")

// Client api responsible for all bot functionality
type BotAPI struct {
	client *Client
}

// Request command to send messages as a bot
type BotMessageCommand struct {
	BotID      string  `json:"bot_id"`
	Message    string  `json:"text"`
	PictureURL *string `json:"picture_url,omitempty"`
}

// Request body to create a bot
type CreateBotCommand struct {
	Name         string  `json:"name"`
	GroupID      string  `json:"group_id"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	CallbackURL  *string `json:"callback_url,omitempty"`
	Notification *bool   `json:"dm_notification,omitempty"`
}

// Request body to update a bot
type UpdateBotCommand struct {
	Name         string  `json:"name"`
	GroupID      string  `json:"group_id"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	CallbackURL  *string `json:"callback_url,omitempty"`
	Notification *bool   `json:"dm_notification,omitempty"`
}

type createBotCommandRequest struct {
	Bot CreateBotCommand `json:"bot"`
}

// Bot data model in GroupMe
type BotDefinitionForGroup struct {
	Name          string  `json:"name"`
	GroupID       string  `json:"group_id"`
	AvatarURL     *string `json:"avatar_url"`
	CallbackURL   *string `json:"callback_url"`
	Notifications bool    `json:"dm_notification"`
	BotID         string  `json:"bot_id"`
}

type bot struct {
	Bot BotDefinitionForGroup `json:"bot"`
}

type deleteBotCommand struct {
	BotID string `json:"bot_id"`
}

// Send message from bot
func (api BotAPI) Send(cmd BotMessageCommand) error {
	url := api.client.makeURL("/v3/bots/post")
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = api.client.getResponse(req)
	if err != nil {
		return err
	}
	return nil
}

func (api BotAPI) Create(cmd CreateBotCommand) (*BotDefinitionForGroup, error) {
	url := api.client.makeURL("/v3/bots")
	envelope := createBotCommandRequest{
		Bot: cmd,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	env := bot{}
	data, err = api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	err = unravel(&data, &env)
	if err != nil {
		return nil, err
	}
	return &env.Bot, nil
}

func (api BotAPI) List() ([]BotDefinitionForGroup, error) {
	url := api.client.makeURL("/v3/bots")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	data, err := api.client.getResponse(req)
	if err != nil {
		return nil, err
	}
	var bots []BotDefinitionForGroup
	err = unravel(&data, &bots)
	if err != nil {
		return nil, err
	}
	return bots, nil
}

func (api BotAPI) Get(botId string) (*BotDefinitionForGroup, error) {
	bots, err := api.List()
	if err != nil {
		return nil, err
	}
	for _, b := range bots {
		if b.BotID == botId {
			return &b, nil
		}
	}
	return nil, ErrBotNotFound
}

// Update simulates updating a bot. GroupMe has no real update endpoint, so
// this is implemented as create-then-delete: a new bot is created with the
// desired fields (falling back to the existing bot's fields for anything not
// explicitly overridden by command), and only once that succeeds is the old
// bot deleted. This avoids leaving the caller with no bot at all if Create
// fails, but it means the returned bot has a NEW BotID distinct from botId.
// Callers must update any stored bot ID and re-register webhooks/callback
// URLs (and any other GroupMe-side configuration keyed on the bot ID) after
// calling this.
//
// hack api until I figure out a better approach with GroupMe apis. Nothing in public docs
func (api BotAPI) Update(botId string, command UpdateBotCommand) (*BotDefinitionForGroup, error) {
	old, err := api.Get(botId)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, fmt.Errorf("no bot exists for botId=%s", botId)
	}

	createCmd := CreateBotCommand(command)
	if createCmd.AvatarURL == nil {
		createCmd.AvatarURL = old.AvatarURL
	}
	if createCmd.CallbackURL == nil {
		createCmd.CallbackURL = old.CallbackURL
	}
	if createCmd.Notification == nil {
		oldNotification := old.Notifications
		createCmd.Notification = &oldNotification
	}

	newBot, err := api.Create(createCmd)
	if err != nil {
		return nil, err
	}

	err = api.Delete(botId)
	if err != nil {
		return nil, err
	}

	return newBot, nil
}

func (api BotAPI) Delete(botId string) error {
	url := api.client.makeURL("/v3/bots/destroy")
	data, err := json.Marshal(deleteBotCommand{
		BotID: botId,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = api.client.getResponse(req)
	if err != nil {
		return err
	}
	return nil
}
