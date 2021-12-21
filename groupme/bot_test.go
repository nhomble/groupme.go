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
