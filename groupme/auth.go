package groupme

import (
	"github.com/nhomble/groupme.go/props"
)

type TokenProvider interface {
	Get() string
}

type SimpleTokenProvider struct {
	token string
}

// Get GroupMe API token
func (p SimpleTokenProvider) Get() string {
	return p.token
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
