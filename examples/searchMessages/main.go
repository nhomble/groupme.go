package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"os"
	"strings"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	provider, err := groupme.TokenPoviderFromProperties(home, ".groupme-api", "prop.txt")
	if err != nil {
		log.Fatal(err)
	}
	client, err := groupme.NewClient(provider)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Messages.Search("", groupme.MessageSearch{
		Criteria: func(message groupme.Message) bool {
			return strings.Contains(strings.ToLower(message.Text), " flu ")
		},
		StopCriteria: func(count int, total int, seen int) bool {
			return seen*5 > total
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range resp.Messages {
		log.Printf("User=%s said=%s at={%v}\n", msg.Name, msg.Text, msg.CreatedAtParsed())
	}
}
