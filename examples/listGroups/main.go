package main

import (
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"os"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	provider, err := groupme.TokenProviderFromProperties(home, ".groupme-api", "prop.txt")
	if err != nil {
		log.Fatal(err)
	}
	client, err := groupme.NewClient(provider)
	if err != nil {
		log.Fatal(err)
	}

	groups, err := client.Groups.FindAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, group := range groups {
		log.Printf("Name=%s ID=%s\n", group.Name, group.ID)
	}
}
