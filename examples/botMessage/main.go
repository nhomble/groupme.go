package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	provider := groupme.EnvironmentTokenProvider{}
	client, err := groupme.NewClient(provider)
	must(err)

	err = client.Bots.Send(groupme.BotMessageCommand{
		BotId:   "your bot id",
		Message: "test",
	})
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
