package listener

import (
	"makini/api"
	"regexp"
)

type BotResponder func(client *api.APIClient, message map[string]interface{}) bool

type BotListener struct {
	Regexp    *regexp.Regexp
	Responder BotResponder
}

var listeners []*BotListener = []*BotListener{}
var UserID string

func Register(re string, responder BotResponder) (*BotListener, error) {
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}

	listener := &BotListener{Regexp: r, Responder: responder}
	listeners = append(listeners, listener)

	return listener, nil
}

func Process(client *api.APIClient, message map[string]interface{}) {
	text := string(message["text"].(string))
	for _, listener := range listeners {
		if listener.Regexp.MatchString(text) {
			if listener.Responder(client, message) {
				return
			}
		}
	}
}

func ProcessMessages(botClient *api.APIClient, in chan *api.APIResponse) {
	for {
		obj := <-in

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != UserID {
						Process(botClient, data)
					}
				}
			}
		}
	}
}
