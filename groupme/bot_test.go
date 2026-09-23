package groupme

import (
	"github.com/jarcoal/httpmock"
	"testing"
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
