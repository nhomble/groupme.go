package groupme

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestSendMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots/post",
		httpmock.NewStringResponder(200, `{}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	if err := client.Bots.Send(BotMessageCommand{
		"botId",
		"Hello",
		nil,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if httpmock.GetTotalCallCount() != 1 {
		t.Errorf("Did not mock send message")
	}
}

func TestBotMessageCommandOmitsUnsetPictureURL(t *testing.T) {
	cmd := BotMessageCommand{
		BotID:   "botId",
		Message: "Hello",
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	marshaled := string(data)

	if strings.Contains(marshaled, `"picture_url"`) {
		t.Errorf("expected marshaled command to omit picture_url, got %s", marshaled)
	}
}

func TestGetBotNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/bots",
		httpmock.NewStringResponder(200, `{"response": [{"bot_id": "other-bot"}]}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	bot, err := client.Bots.Get("missing-bot")

	if bot != nil {
		t.Errorf("Expected nil bot, got %v", bot)
	}
	if err != ErrBotNotFound {
		t.Errorf("Expected ErrBotNotFound, got %v", err)
	}
}

func TestUpdatePreservesUnspecifiedFields(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/bots",
		httpmock.NewStringResponder(200, `{"response": [{
			"bot_id": "old-bot",
			"name": "OldName",
			"group_id": "group-1",
			"avatar_url": "http://example.com/avatar.png",
			"callback_url": "http://example.com/callback",
			"dm_notification": true
		}]}`))

	var createBody createBotCommandRequest
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots",
		func(req *http.Request) (*http.Response, error) {
			data, _ := io.ReadAll(req.Body)
			if err := json.Unmarshal(data, &createBody); err != nil {
				t.Fatalf("failed to unmarshal create body: %v", err)
			}
			return httpmock.NewStringResponse(200, `{"response": {"bot": {"bot_id": "new-bot"}}}`), nil
		})

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots/destroy",
		httpmock.NewStringResponder(200, `{}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	newBot, err := client.Bots.Update("old-bot", UpdateBotCommand{
		Name:    "NewName",
		GroupID: "group-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newBot.BotID != "new-bot" {
		t.Errorf("expected new bot id, got %s", newBot.BotID)
	}

	if createBody.Bot.AvatarURL == nil || *createBody.Bot.AvatarURL != "http://example.com/avatar.png" {
		t.Errorf("expected avatar_url preserved, got %v", createBody.Bot.AvatarURL)
	}
	if createBody.Bot.CallbackURL == nil || *createBody.Bot.CallbackURL != "http://example.com/callback" {
		t.Errorf("expected callback_url preserved, got %v", createBody.Bot.CallbackURL)
	}
	if createBody.Bot.Notification == nil || *createBody.Bot.Notification != true {
		t.Errorf("expected dm_notification preserved as true, got %v", createBody.Bot.Notification)
	}
}

func TestUpdateCreatesBeforeDeleting(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/bots",
		httpmock.NewStringResponder(200, `{"response": [{"bot_id": "old-bot", "name": "OldName", "group_id": "group-1"}]}`))

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots",
		httpmock.NewStringResponder(500, `{"meta": {"errors": ["boom"]}}`))

	deleteCalled := false
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots/destroy",
		func(req *http.Request) (*http.Response, error) {
			deleteCalled = true
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Bots.Update("old-bot", UpdateBotCommand{
		Name:    "NewName",
		GroupID: "group-1",
	})

	if err == nil {
		t.Fatalf("expected error when create fails")
	}
	if deleteCalled {
		t.Errorf("expected Delete not to be called when Create fails")
	}
}
