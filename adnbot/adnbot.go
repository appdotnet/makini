package main

import (
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"makini/api"
	"makini/listener"
	"makini/stream"
)

var (
	file = flag.String("config", "config.yaml", "YAML config file")
)

func main() {
	flag.Parse()

	config, err := yaml.ReadFile(*file)
	if err != nil {
		log.Fatalf("Error loading config (%q): %s", *file, err)
	}

	userToken, _ := config.Get("tokens.user")
	appToken, _ := config.Get("tokens.app")

	userClient := &api.APIClient{AccessToken: userToken}
	appClient := &api.APIClient{AccessToken: appToken}

	url := appClient.GetStreamEndpoint("makini")
	listener.UserID = userClient.GetUserID()

	_, err = listener.Register("^send invite to ([-.+_a-zA-Z0-9@]+)$", func(message *listener.BotMessage) bool {
		message.Reply(fmt.Sprintf("OK, I'll send an invite to %s.", message.Matches[1]))
		return false
	})

	messages := stream.ProcessStream(url)
	listener.ProcessMessages(userClient, messages)
}
