package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"testing"
	"time"
)

func TestGetUser(t *testing.T) {
	client := AClient()
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
	client := AClient()
	newName := RandomName()
	user, err := client.Users.Get()
	if err != nil {
		t.Fatal(err)
	}
	for newName == user.Name {
		newName = RandomName()
	}

	update := &groupme.UpdateUserCommand{
		Name:  &newName,
		Email: &user.Email,
	}
	_, err = client.Users.Update(update)
	if err != nil {
		t.Fatal(err)
	}

	await(t, 1*time.Second, 10*time.Second, func() bool {
		updated, err := client.Users.Get()
		if err != nil {
			t.Fatal(err)
		}
		return user.Name == updated.Name
	})
}
