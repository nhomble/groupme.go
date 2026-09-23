package groupme

import (
	"strings"
	"testing"
)

func TestUnravelMissingResponse(t *testing.T) {
	data := []byte(`{"meta":{"code":200}}`)

	var dest map[string]interface{}
	err := unravel(&data, &dest)

	if err == nil {
		t.Fatal("expected error for missing \"response\" field, got nil")
	}
}

func TestUnravelNullResponse(t *testing.T) {
	data := []byte(`{"response":null,"meta":{"code":200}}`)

	var dest map[string]interface{}
	err := unravel(&data, &dest)

	if err == nil {
		t.Fatal("expected error for null \"response\" field, got nil")
	}
}

func TestUnravelValidResponse(t *testing.T) {
	data := []byte(`{"response":{"id":"123","name":"test"},"meta":{"code":200}}`)

	var dest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := unravel(&data, &dest)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if dest.ID != "123" || dest.Name != "test" {
		t.Errorf("unexpected unravel result: %+v", dest)
	}
}

func TestParseErrorMissingMeta(t *testing.T) {
	data := []byte(`{"response":null}`)

	msg := parseError(&data)

	if !strings.Contains(msg, "meta") {
		t.Errorf("expected error message to mention missing meta field, got %q", msg)
	}
}

func TestParseErrorNullMeta(t *testing.T) {
	data := []byte(`{"response":null,"meta":null}`)

	msg := parseError(&data)

	if !strings.Contains(msg, "meta") {
		t.Errorf("expected error message to mention missing meta field, got %q", msg)
	}
}

func TestParseErrorMissingErrors(t *testing.T) {
	data := []byte(`{"meta":{"code":400}}`)

	msg := parseError(&data)

	if !strings.Contains(msg, "errors") {
		t.Errorf("expected error message to mention missing errors field, got %q", msg)
	}
}

func TestParseErrorValid(t *testing.T) {
	data := []byte(`{"meta":{"code":400,"errors":["invalid token","bad request"]}}`)

	msg := parseError(&data)

	if msg != "invalid token\nbad request" {
		t.Errorf("unexpected parseError result: %q", msg)
	}
}
