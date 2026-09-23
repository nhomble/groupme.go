package groupme

import (
	"errors"
	"github.com/nhomble/groupme.go/props"
	"os"
	"path/filepath"
)

type TokenProvider interface {
	Get() (string, error)
}

type SimpleTokenProvider struct {
	token string
}

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

// Create token provider from in memory token
func TokenProviderFromToken(t string) TokenProvider {
	return SimpleTokenProvider{token: t}
}

// Create token provider from properties file
func TokenProviderFromProperties(p ...string) (TokenProvider, error) {
	thePath := filepath.Join(p...)
	config, err := props.View(thePath)
	if err != nil {
		return nil, err
	}
	return SimpleTokenProvider{token: config["token"]}, nil
}
