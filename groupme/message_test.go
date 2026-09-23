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

// registerMessagesResponder registers a GET responder for the messages
// endpoint that always returns the given canned JSON body.
func registerMessagesResponder(body string) {
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, body), nil
		})
}

func TestSearchZeroValueDoesNotPanic(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	registerMessagesResponder(`{"response":{"count":0,"messages":[]}}`)

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

func TestQueryMessagesQueryStringWellFormed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedRawQuery string
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			capturedRawQuery = req.URL.RawQuery
			return httpmock.NewStringResponse(200, `{"response":{"count":0,"messages":[]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	beforeId := "abc"
	_, err := client.Messages.Query("groupId", &MessageQuery{BeforeId: &beforeId})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if strings.Contains(capturedRawQuery, "?&") || strings.HasPrefix(capturedRawQuery, "&") {
		t.Errorf("expected well-formed query string, got %q", capturedRawQuery)
	}
	if !strings.Contains(capturedRawQuery, "before_id=abc") {
		t.Errorf("expected before_id=abc in query, got %q", capturedRawQuery)
	}
}

func TestQueryMessagesEscapesGroupIDAndIDs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedPath, capturedRawQuery string
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/.*`,
		func(req *http.Request) (*http.Response, error) {
			capturedPath = req.URL.EscapedPath()
			capturedRawQuery = req.URL.RawQuery
			return httpmock.NewStringResponse(200, `{"response":{"count":0,"messages":[]}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	beforeId := "id/with space"
	_, err := client.Messages.Query("group/id", &MessageQuery{BeforeId: &beforeId})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedPath := "/v3/groups/group%2Fid/messages"
	if capturedPath != expectedPath {
		t.Errorf("expected escaped path %q, got %q", expectedPath, capturedPath)
	}
	if !strings.Contains(capturedRawQuery, "before_id=id%2Fwith+space") {
		t.Errorf("expected before_id to be encoded, got %q", capturedRawQuery)
	}
}

func TestSendAcceptsValueConstructedCommand(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", `=~^https://api\.groupme\.com/v3/groups/groupId/messages`,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"response":{"message":{"id":"1","text":"hello"}}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))

	// SendMessageCommand is a value type, so there is no nil-pointer panic
	// risk when constructing and passing it directly.
	cmd := SendMessageCommand{
		SourceGuid: "guid-1",
		Text:       "hello",
	}
	message, err := client.Messages.Send("groupId", cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if message == nil || message.Text != "hello" {
		t.Errorf("expected sent message with text 'hello', got %+v", message)
	}
}

func TestSearchRespectsLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	registerMessagesResponder(`{"response":{"count":10,"messages":[{"id":"1"},{"id":"2"},{"id":"3"}]}}`)

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
