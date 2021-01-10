package main

import (
	"fmt"
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
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
		SourceGuid: fmt.Sprintf("%d%d", rand.Int63(), rand.Int63()),
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
