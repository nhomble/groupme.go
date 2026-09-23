package groupme

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestUpdateUserCommandOmitsUnsetFields(t *testing.T) {
	name := "new-name"
	cmd := UpdateUserCommand{
		Name: &name,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	marshaled := string(data)

	if !strings.Contains(marshaled, `"name":"new-name"`) {
		t.Errorf("expected marshaled command to include name, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"email"`) {
		t.Errorf("expected marshaled command to omit email, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"zip_code"`) {
		t.Errorf("expected marshaled command to omit zip_code, got %s", marshaled)
	}
	if strings.Contains(marshaled, `"avatar_url"`) {
		t.Errorf("expected marshaled command to omit avatar_url, got %s", marshaled)
	}
}

func TestGetUser(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var capturedMethod, capturedPath string
	httpmock.RegisterResponder("GET", "https://api.groupme.com/v3/users/me",
		func(req *http.Request) (*http.Response, error) {
			capturedMethod = req.Method
			capturedPath = req.URL.Path
			return httpmock.NewStringResponse(200, `{"response":{"id":"1","name":"Test User","email":"test@example.com"}}`), nil
		})

	client, _ := NewClient(TokenProviderFromToken("test"))
	user, err := client.Users.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != "GET" {
		t.Errorf("expected GET, got %s", capturedMethod)
	}
	if capturedPath != "/v3/users/me" {
		t.Errorf("expected path /v3/users/me, got %s", capturedPath)
	}
	if user == nil || user.ID != "1" || user.Name != "Test User" {
		t.Errorf("unexpected user result: %+v", user)
	}
}

func TestUpdateDoesNotPrintPIIToStdout(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://api.groupme.com/v3/users/update",
		httpmock.NewStringResponder(200, `{"response":{"id":"1","name":"Test User","email":"test@example.com"}}`))

	client, _ := NewClient(TokenProviderFromToken("test"))

	name := "Test User"
	email := "test@example.com"
	zip := "12345"
	cmd := UpdateUserCommand{
		Name:    &name,
		Email:   &email,
		ZipCode: &zip,
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	user, updateErr := client.Users.Update(cmd)

	w.Close()
	os.Stdout = origStdout

	var buf strings.Builder
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Fatalf("failed to read captured stdout: %v", copyErr)
	}

	if updateErr != nil {
		t.Fatalf("unexpected error: %v", updateErr)
	}
	if user == nil || user.Name != "Test User" {
		t.Fatalf("expected updated user with name 'Test User', got %+v", user)
	}

	captured := buf.String()
	if captured != "" {
		t.Errorf("expected no stdout output, got: %q", captured)
	}
	if strings.Contains(captured, email) || strings.Contains(captured, zip) {
		t.Errorf("stdout leaked PII: %q", captured)
	}
}
