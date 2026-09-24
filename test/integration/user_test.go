//go:build integration

package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"os"
	"testing"
	"time"
)

func TestGetUser(t *testing.T) {
	client := AClient()
	user, err := client.Users.Get()
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == "" {
		t.Errorf("Expected non-empty User.ID")
	}
	if expected := os.Getenv("GROUPME_USER_ID"); expected != "" && user.ID != expected {
		t.Errorf("User.ID | %s!=%s", expected, user.ID)
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

	update := groupme.UpdateUserCommand{
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
			return false
		}
		return updated.Name == newName
	})
}
