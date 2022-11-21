package main

import (
	"log"

	"github.com/nhomble/groupme.go/groupme"
)

func main() {
	provider := groupme.EnvironmentTokenProvider{}
	client, _ := groupme.NewClient(provider)
	err := client.Bots.Delete("foo")
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
