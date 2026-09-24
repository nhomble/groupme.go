package groupme

import (
	"errors"
	"github.com/nhomble/groupme.go/props"
	"os"
	"path/filepath"
)

// TokenProvider supplies the GroupMe API token used to authenticate requests.
type TokenProvider interface {
	Get() (string, error)
}

// SimpleTokenProvider provides a fixed, in-memory token.
type SimpleTokenProvider struct {
	token string
}

// EnvironmentTokenProvider provides a token read from an environment variable.
type EnvironmentTokenProvider struct {
	Key string // optional field that is the environment variable key
}

// Get GroupMe API token
func (p SimpleTokenProvider) Get() (string, error) {
	if len(p.token) == 0 {
		return "", errors.New("token is empty")
	}
	return p.token, nil
}

// Get GroupMe API token from environment
func (e EnvironmentTokenProvider) Get() (string, error) {
	k := "GO_GROUPME_API_TOKEN"
	if len(e.Key) > 0 {
		k = e.Key
	}
	t := os.Getenv(k)
	if len(t) == 0 {
		return "", errors.New("token is empty")
	}
	return t, nil
}

// TokenProviderFromToken creates a token provider from an in-memory token.
func TokenProviderFromToken(t string) TokenProvider {
	return SimpleTokenProvider{token: t}
}

// TokenProviderFromProperties creates a token provider from a properties
// file, joining the given path segments.
func TokenProviderFromProperties(p ...string) (TokenProvider, error) {
	thePath := filepath.Join(p...)
	config, err := props.View(thePath)
	if err != nil {
		return nil, err
	}
	return SimpleTokenProvider{token: config["token"]}, nil
}
