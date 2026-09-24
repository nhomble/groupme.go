//go:build integration

package main

import (
	"fmt"
	"github.com/nhomble/groupme.go/groupme"
	"log"
	"math/rand"
	"time"
)

func init() {
	// go.mod targets go 1.15; math/rand is not auto-seeded before go 1.20,
	// so seed explicitly to avoid deterministic "random" names across runs.
	rand.Seed(time.Now().UnixNano())
}

func AClient() *groupme.Client {
	// configured in github secret settings
	provider := groupme.EnvironmentTokenProvider{Key: "GROUPME_TOKEN"}
	client, err := groupme.NewClient(provider)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func RandomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	s := ""
	for i := 0; i < 5+(rand.Int()%10); i++ {
		s += string(chars[rand.Int()%len(chars)])
	}
	return fmt.Sprintf("super test %s", s)
}
