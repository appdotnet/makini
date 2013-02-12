package main

import (
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"makini/api"
	"makini/stream"
)

var (
	file = flag.String("config", "config.yaml", "YAML config file")
)

var userID string

func processMessage(botClient *api.APIClient, in chan *api.APIResponse) {
	for {
		obj := <-in

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != userID {
						log.Print("Got message: ", data["text"], " from ", user["username"])
						msg := fmt.Sprintf("Hi, @%s! What's up?", user["username"])
						botClient.Reply(data["channel_id"].(string), msg)
					}
				}
			}
		}
	}
}

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
	userID = userClient.GetUserID()

	messages := stream.ProcessStream(url)
	processMessage(userClient, messages)
}
