package listener

import (
	"log"
	"makini/api"
	"regexp"
	"strings"
)

var listeners []*BotListener = []*BotListener{}
var UserID string

type BotResponder func(message *BotMessage) bool

type BotListener struct {
	Regexp    *regexp.Regexp
	Responder BotResponder
}

type BotMessage struct {
	Message   map[string]interface{}
	BotClient *api.APIClient
	Matches   []string
	Text      string
	channelID string
}

func (message *BotMessage) Reply(text string) {
	contents := map[string]interface{}{
		"text": text,
	}

	message.ReplyJSON(contents)
}

func (message *BotMessage) ReplyJSON(contents map[string]interface{}) {
	message.BotClient.Reply(message.channelID, contents)
}

func (message *BotMessage) Process() {
	for _, listener := range listeners {
		if matches := listener.Regexp.FindStringSubmatch(message.Text); matches != nil {
			message.Matches = matches
			if listener.Responder(message) {
				return
			}
		}
	}
}

func Register(re string, responder BotResponder) (*BotListener, error) {
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}

	listener := &BotListener{Regexp: r, Responder: responder}
	listeners = append(listeners, listener)

	return listener, nil
}

func ProcessMessages(botClient *api.APIClient, in chan *api.APIResponse) {
	for {
		obj := <-in

		log.Printf("Got message: %s", obj)

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != UserID {
						message := &BotMessage{
							Message:   data,
							BotClient: botClient,
							Text:      strings.TrimSpace(data["text"].(string)),
							channelID: data["channel_id"].(string),
						}
						message.Process()
					}
				}
			}
		}
	}
}
