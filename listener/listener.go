package listener

import (
	"makini/api"
	"regexp"
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
	channelID string
}

func (message *BotMessage) Text() string {
	return message.Message["text"].(string)
}

func (message *BotMessage) Reply(text string) {
	message.BotClient.Reply(message.channelID, text)
}

func (message *BotMessage) Process() {
	text := message.Text()
	for _, listener := range listeners {
		if matches := listener.Regexp.FindStringSubmatch(text); matches != nil {
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

		if obj.Meta["type"] == "message" && obj.Meta["channel_type"] == "net.app.core.pm" {
			if data, ok := obj.Data.(map[string]interface{}); ok {
				if user, ok := data["user"].(map[string]interface{}); ok {
					if user["id"] != UserID {
						message := &BotMessage{
							Message: data,
							BotClient: botClient,
							channelID: data["channel_id"].(string),
						}
						message.Process()
					}
				}
			}
		}
	}
}
