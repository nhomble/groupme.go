package groupme

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestQueryMessagesLimitDereferenced(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			if !strings.Contains(req.URL.String(), "limit=50") {
				t.Errorf("expected URL to contain limit=50, got %s", req.URL.String())
			}
			return httpmock.NewStringResponse(200, `{"response":{"count":0,"messages":[]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	limit := 50
	_, err := client.Messages.Query("groupId", &MessageQuery{Limit: &limit})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if httpmock.GetTotalCallCount() != 1 {
		t.Errorf("Did not mock query messages")
	}
}

func TestQueryMessagesLimitAboveMaxRejected(t *testing.T) {
	client, _ := NewClient(TokenProviderFromToken("test"))
	limit := 101
	_, err := client.Messages.Query("groupId", &MessageQuery{Limit: &limit})
	if err == nil {
		t.Errorf("expected error for limit above 100, got nil")
	}
}
