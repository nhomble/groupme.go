package groupme

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestGetResponse401ReturnsAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		httpmock.NewStringResponder(401, `{"meta":{"code":401,"errors":["invalid token"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Users.Get()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected StatusCode=401, got %d", apiErr.StatusCode)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0] != "invalid token" {
		t.Errorf("expected parsed errors [invalid token], got %v", apiErr.Errors)
	}
}

func TestGetResponse404UnparseableBodyStillSetsStatusCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		httpmock.NewStringResponder(404, `not json`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Users.Get()

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
	if len(apiErr.Errors) != 0 {
		t.Errorf("expected no parsed errors for unparseable body, got %v", apiErr.Errors)
	}
}

// errReader is an io.Reader that always fails, used to simulate a body read
// failure (e.g. a connection dropped mid-read).
type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated read failure")
}

func TestGetResponseReadAllFailureSurfacesError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(errReader{}),
				Header:     make(http.Header),
			}
			return resp, nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Users.Get()

	if err == nil {
		t.Fatal("expected error from failed read, got nil")
	}
	if _, ok := err.(*APIError); ok {
		t.Errorf("expected a plain read error, not an *APIError, got %v", err)
	}
}
