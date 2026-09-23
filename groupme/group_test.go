package groupme

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestJoinEscapesShareUrlAndGroupId(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedPath string
	httpmock.RegisterResponder("POST", `=~^https://api\.groupme\.com/v3/groups/.*`,
		func(req *http.Request) (*http.Response, error) {
			capturedPath = req.URL.EscapedPath()
			return httpmock.NewStringResponse(200, `{"response":{}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Groups.Join("group/id", "share url/with space")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "/v3/groups/group%2Fid/join/share%20url%2Fwith%20space"
	if capturedPath != expected {
		t.Errorf("expected escaped path %q, got %q", expected, capturedPath)
	}

	if httpmock.GetTotalCallCount() != 1 {
		t.Errorf("Did not mock join group")
	}
}

func TestGetGroupEscapesId(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedPath string
	httpmock.RegisterResponder("GET", `=~^https://api\.groupme\.com/v3/groups/.*`,
		func(req *http.Request) (*http.Response, error) {
			capturedPath = req.URL.EscapedPath()
			return httpmock.NewStringResponse(200, `{"response":{}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Groups.Get("weird/id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "/v3/groups/weird%2Fid"
	if capturedPath != expected {
		t.Errorf("expected escaped path %q, got %q", expected, capturedPath)
	}
}
