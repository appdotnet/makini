package main

import (
	"flag"
	"fmt"
	"log"
	"makini/api"
	"makini/stream"
	"github.com/kylelemons/go-gypsy/yaml"
)

var (
	file = flag.String("config", "config.yaml", "YAML config file")
)

var userID string

func logStream(userClient *api.APIClient, in chan *api.APIResponse) {
	for {
		obj := <-in

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != userID {
						log.Print("Got message: ", data["text"], " from ", user["username"])
						msg := fmt.Sprintf("Hi, @%s! What's up?", user["username"])
						userClient.Reply(data["channel_id"].(string), msg)
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

	bytes := make(chan []byte)
	go stream.ConsumeStream(url, bytes)

	messages := make(chan *api.APIResponse)
	go stream.UnmarshalStream(bytes, messages)

	logStream(userClient, messages)
}
