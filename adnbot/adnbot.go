package main

import (
	"flag"
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

	_, err = listener.Register("^[a-z]+$", func(client *api.APIClient, message map[string]interface{}) bool {
		client.Reply(message["channel_id"].(string), "Hey")
		return false
	})

	_, err = listener.Register("^[a-z]+$", func(client *api.APIClient, message map[string]interface{}) bool {
		log.Print("AYO!")
		return true
	})

	messages := stream.ProcessStream(url)
	listener.ProcessMessages(userClient, messages)
}
