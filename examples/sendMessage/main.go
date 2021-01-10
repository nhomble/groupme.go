package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"os"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	provider, err := groupme.TokenPoviderFromProperties(home + "/.groupme-go.prop")
	if err != nil {
		log.Fatal(err)
	}
	client, err := groupme.NewClient(provider)

	if err != nil {
		log.Fatal(err)
	}
	group, err := client.Groups.Create(&groupme.CreateGroupCommand{
		Name:  "hombro-test",
		Share: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Group created id=%s name=%s\n", group.Id, group.Name)
	_, err = client.Messages.Send(group.Id, &groupme.SendMessageCommand{
		SourceGuid: uuid.New().String(),
		Text:       "Message sent!",
	})

	if err != nil {
		client.Groups.Delete(group.Id)
		log.Fatal(err)
	}
	messages, err := client.Messages.Query(group.Id, nil)
	if err != nil {
		client.Groups.Delete(group.Id)
		log.Fatal(err)
	}
	fmt.Printf("Text in chat.... '%s'\n", messages.Messages[0].Text)
	client.Groups.Delete(group.Id)
}
