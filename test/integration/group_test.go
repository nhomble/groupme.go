package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"testing"
	"time"
)

func TestFindMakeDeleteGroupt(t *testing.T) {
	client := AClient()
	name := "hombro-test-" + RandomName()
	groups, err := client.Groups.FindAll()
	if err != nil {
		t.Error(err)
	}
	for _, g := range groups {
		if g.Name == name {
			t.Logf("Delete from previous test %v\n", g)
			client.Groups.Delete(g.Id)
		}
	}

	originalNumber := len(groups)
	t.Logf("name=%s originalNumber=%d\n", name, originalNumber)

	result, err := client.Groups.Create(&groupme.CreateGroupCommand{
		Name:  name,
		Share: false,
	})

	if err != nil {
		t.Error(err)
	}
	if result.Name != name {
		t.Errorf("Expected group with name=%s but got %s", name, result.Name)
	}
	await(t, 1*time.Second, 10*time.Second, func() bool {
		groups, err = client.Groups.FindAll()
		if err != nil {
			t.Error(err)
		}
		return len(groups) == (1 + originalNumber)
	})

	groups, err = client.Groups.FindAll()
	if err != nil {
		t.Error(err)
	}
	for i, g := range groups {
		t.Logf("%d> id=%s %s\n", i, g.Id, g.Name)
	}

	err = client.Groups.Delete(result.Id)
	if err != nil {
		t.Error(err)
	}

	await(t, 1*time.Second, 10*time.Second, func() bool {
		groups, err = client.Groups.FindAll()
		if err != nil {
			t.Error(err)
		}
		return len(groups) == originalNumber
	})
	groups, err = client.Groups.FindAll()
	if err != nil {
		t.Error(err)
	}
	for i, g := range groups {
		t.Logf("%d> id=%s name=%s\n", i, g.Id, g.Name)
	}
}
