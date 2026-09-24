package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"log"
)

func main() {
	provider := groupme.EnvironmentTokenProvider{}
	client, err := groupme.NewClient(provider)
	must(err)

	err = client.Bots.Send(groupme.BotMessageCommand{
		BotID:   "your bot id",
		Message: "test",
	})
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
