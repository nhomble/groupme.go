package groupme

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

// assertAPIError asserts that err is a non-nil *APIError with the given
// StatusCode and Errors, and returns it for further inspection.
func assertAPIError(t *testing.T, err error, wantStatus int, wantErrors []string) *APIError {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != wantStatus {
		t.Errorf("expected StatusCode=%d, got %d", wantStatus, apiErr.StatusCode)
	}
	if !reflect.DeepEqual(apiErr.Errors, wantErrors) {
		t.Errorf("expected parsed errors %v, got %v", wantErrors, apiErr.Errors)
	}
	return apiErr
}

func TestGetResponse401ReturnsAPIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		httpmock.NewStringResponder(401, `{"meta":{"code":401,"errors":["invalid token"]}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Users.Get()

	assertAPIError(t, err, 401, []string{"invalid token"})
}

func TestGetResponse404UnparseableBodyStillSetsStatusCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		httpmock.NewStringResponder(404, `not json`))

	client, _ := NewClient(TokenProviderFromToken("test"))
	_, err := client.Users.Get()

	assertAPIError(t, err, 404, nil)
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
