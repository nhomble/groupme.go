package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/nhomble/groupme.go/groupme"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	provider := groupme.EnvironmentTokenProvider{}
	client, err := groupme.NewClient(provider)
	must(err)

	list, err := client.Bots.List()
	log.Printf("%v\n", list)
	s := "localhost:1234"
	_, err = client.Bots.Create(groupme.CreateBotCommand{
		Name:        "test",
		GroupId:     "11617071",
		CallbackUrl: &s,
	})
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
