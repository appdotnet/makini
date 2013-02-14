package main

import (
	"encoding/json"
	"flag"
	"log"
	"makini/api"
	"makini/listener"
	"makini/stream"
	"os"
)

type Config struct {
	ADN struct {
		TokenURLBase      string `json:"token_url_base"`
		TokenHostOverride string `json:"token_host_override"`
		APIURLBase        string `json:"api_url_base"`
		APIHostOverride   string `json:"api_host_override"`
		StreamURLOverride string `json:"stream_url_override"`
		ClientID          string `json:"client_id"`
		ClientSecret      string `json:"client_secret"`
		UserID            string `json:"user_id"`
		StreamKey         string `json:"stream_key"`
	} `json:"adn"`
}

var (
	file = flag.String("config", "config.json", "JSON config file")
)

func main() {
	flag.Parse()

	file, err := os.Open(*file)
	if err != nil {
		log.Fatalf("Error loading config (%q): %s", *file, err)
	}

	decoder := json.NewDecoder(file)
	var config Config
	if err = decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config: %s", err)
	}

	api.OAuthURLBase = config.ADN.OAuthURLBase
	api.APIURLBase = config.ADN.APIURLBase
	api.ClientID = config.ADN.ClientID
	api.ClientSecret = config.ADN.ClientSecret

	userClient, err := api.GetToken(map[string]string{
		"grant_type": "xyx_mxml_internal_implicit_token",
		"user_id":    config.ADN.UserID,
		"scopes":     "messages",
	})

	userClient := &api.APIClient{AccessToken: config.ADN.UserToken}
	listener.UserID = userClient.GetUserID()

	appClient, err := api.GetToken(map[string]string{
		"grant_type": "client_credentials",
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Print(appClient)

	url := appClient.GetStreamEndpoint(config.ADN.StreamKey)

	log.Print("SUP:", url)

	messages := stream.ProcessStream(url)
	listener.ProcessMessages(userClient, messages)
}
