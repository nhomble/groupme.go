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

func TestSearchZeroValueDoesNotPanic(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"response":{"count":0,"messages":[]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	result, err := client.Messages.Search("groupId", MessageSearch{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || len(result.Messages) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestSearchStopsOnEmptyPage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	calls := 0
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			calls += 1
			if calls == 1 {
				return httpmock.NewStringResponse(200, `{"response":{"count":5,"messages":[{"id":"1"},{"id":"2"}]}}`), nil
			}
			return httpmock.NewStringResponse(200, `{"response":{"count":5,"messages":[]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	result, err := client.Messages.Search("groupId", MessageSearch{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result.Messages))
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestSearchRespectsLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"response":{"count":10,"messages":[{"id":"1"},{"id":"2"},{"id":"3"}]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	limit := 2
	result, err := client.Messages.Search("groupId", MessageSearch{Limit: &limit})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result.Messages) != limit {
		t.Errorf("expected %d messages, got %d", limit, len(result.Messages))
	}
}
