package groupme

import (
	"github.com/nhomble/groupme.go/props"
	"os"
)

type TokenProvider interface {
	Get() string
}

type SimpleTokenProvider struct {
	token string
}

type EnvironmentTokenProvider struct {
	Key string // optional field that is the environment variable key
}

// Get GroupMe API token
func (p SimpleTokenProvider) Get() string {
	return p.token
}

// Get GroupMe API token from environment
func (e EnvironmentTokenProvider) Get() string {
	k := "GO_GROUPME_API_TOKEN"
	if len(e.Key) > 0 {
		k = e.Key
	}
	return os.Getenv(k)
}

// Create token provider from in memory token
func TokenProviderFromToken(t string) TokenProvider {
	return SimpleTokenProvider{token: t}
}

// Create token provider from properties file
func TokenPoviderFromProperties(p string) (TokenProvider, error) {
	config, err := props.View(p)
	if err != nil {
		return nil, err
	}
	return SimpleTokenProvider{token: (*config)["token"]}, nil
}
