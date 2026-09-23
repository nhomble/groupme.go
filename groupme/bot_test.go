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
	client.Bots.Send(BotMessageCommand{
		"botId",
		"Hello",
		nil,
	})

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

func TestListBots(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/bots",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			return httpmock.NewStringResponse(200, `{"response":[{"bot_id":"bot-1","name":"Bot One"},{"bot_id":"bot-2","name":"Bot Two"}]}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	bots, err := client.Bots.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "GET" {
		t.Errorf("expected GET, got %s", capturedMethod)
	}
	if capturedPath != "/v3/bots" {
		t.Errorf("expected path /v3/bots, got %s", capturedPath)
	}
	if len(bots) != 2 || bots[0].BotID != "bot-1" || bots[1].BotID != "bot-2" {
		t.Errorf("unexpected bots result: %+v", bots)
	}
}

func TestListBotsAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/bots",
		httpmock.NewStringResponder(500, `{"meta":{"errors":["boom"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	bots, err := client.Bots.List()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if bots != nil {
		t.Errorf("expected nil bots on error, got %v", bots)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected StatusCode=500, got %d", apiErr.StatusCode)
	}
}

func TestDeleteBot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	var capturedBody deleteBotCommand
	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots/destroy",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			data, _ := io.ReadAll(req.Body)
			if err := json.Unmarshal(data, &capturedBody); err != nil {
				t.Fatalf("failed to unmarshal delete body: %v", err)
			}
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	if err := client.Bots.Delete("bot-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected POST, got %s", capturedMethod)
	}
	if capturedPath != "/v3/bots/destroy" {
		t.Errorf("expected path /v3/bots/destroy, got %s", capturedPath)
	}
	if capturedBody.BotID != "bot-1" {
		t.Errorf("expected bot_id=bot-1 in body, got %q", capturedBody.BotID)
	}
}

func TestDeleteBotAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/bots/destroy",
		httpmock.NewStringResponder(404, `{"meta":{"errors":["not found"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	err := client.Bots.Delete("bot-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected StatusCode=404, got %d", apiErr.StatusCode)
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
