package main

import (
	"fmt"
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"math/rand"
	"testing"
)

func randomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	s := ""
	for i := 0; i < 5+(rand.Int()%10); i++ {
		s += string(chars[rand.Int()%len(chars)])
	}
	return fmt.Sprintf("super test %s", s)
}

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

func TestGetUser(t *testing.T) {
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

func TestUpdateName(t *testing.T) {
	client := client()
	newName := randomName()
	user, err := client.Users.Get()
	if err != nil {
		t.Fatal(err)
	}
	for newName == user.Name {
		newName = randomName()
	}

	update := &groupme.UpdateUserCommand{
		Name:  &newName,
		Email: &user.Email,
	}
	_, err = client.Users.Update(update)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := client.Users.Get()
	if err != nil {
		t.Fatal(err)
	}
	if user.Name == updated.Name {
		t.Fatal("Name didn't update!")
	}
}
