package main

import (
	"fmt"
	"log"

	"github.com/nhomble/groupme.go/groupme"
)

func main() {
	provider := groupme.EnvironmentTokenProvider{}
	client, err := groupme.NewClient(provider)
	must(err)

	list, _ := client.Bots.List()
	for _, b := range list {
		fmt.Printf("%s %s %s\n", b.Name, b.BotId, b.GroupId)
	}

	avatarURL := "https://imagehost.com/avatar.jpg"
	callBackURL := "http://null.com/a"
	bot, err := client.Bots.Create(groupme.CreateBotCommand{
		Name:        "test",
		GroupId:     "11617071",
		AvatarURL:   &avatarURL,
		CallbackURL: &callBackURL,
	})
	must(err)

	bot2, _ := client.Bots.Get((*bot).BotId)
	client.Bots.Update(bot2.BotId, groupme.UpdateBotCommand{
		Name:        "test2",
		GroupId:     "11617071",
		AvatarURL:   &avatarURL,
		CallbackURL: &callBackURL,
	})
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
