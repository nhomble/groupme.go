package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"testing"
)

func client() *groupme.Client {
	// configured in github secret settings
	provider := groupme.EnvironmentTokenProvider{Key: "GROUPME_TOKEN"}
	log.Println(provider.Get())
	client, err := groupme.NewClient(provider, nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func TestUser(t *testing.T) {
	client := client()
	user, err := client.Users.Get()
	if err != nil {
		t.Fatal(err)
	}
	expected := "23807192"
	if user.Id != expected {
		t.Errorf("User.Id | %s!=%s", expected, user.Id)
	}
}
