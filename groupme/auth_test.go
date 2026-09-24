package groupme

import (
	"testing"
)

func TestSimpleTokenProviderEmptyTokenReturnsError(t *testing.T) {
	provider := TokenProviderFromToken("")
	_, err := provider.Get()
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestEnvironmentTokenProviderEmptyTokenReturnsError(t *testing.T) {
	key := "GROUPME_TOKEN_EMPTY_TEST"
	t.Setenv(key, "")
	provider := EnvironmentTokenProvider{Key: key}
	_, err := provider.Get()
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestNewClientEmptyTokenReturnsError(t *testing.T) {
	client, err := NewClient(TokenProviderFromToken(""))
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if client != nil {
		t.Fatal("expected nil client when token is empty")
	}
}

func TestNewClientValidTokenSucceeds(t *testing.T) {
	client, err := NewClient(TokenProviderFromToken("test"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
