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

	list, err := client.Bots.List()
	must(err)
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

	bot2, err := client.Bots.Get(bot.BotId)
	must(err)
	_, err = client.Bots.Update(bot2.BotId, groupme.UpdateBotCommand{
		Name:        "test2",
		GroupId:     "11617071",
		AvatarURL:   &avatarURL,
		CallbackURL: &callBackURL,
	})
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
