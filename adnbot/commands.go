package main

import (
	"log"
	"fmt"
	"makini/listener"
)

var annotation = []interface{}{
	map[string]interface{}{
		"type": "net.app.core.oembed",
		"value": map[string]interface{}{
			"author_name":      "@fooby",
			"author_url":       "https://alpha.app.net/fooby",
			"height":           200,
			"provider_name":    "App.net",
			"provider_url":     "https://app.net",
			"thumbnail_height": 200,
			"thumbnail_url":    "http://i.imgur.com/UmpOi.gif",
			"thumbnail_width":  200,
			"title":            "UmpOi.gif",
			"type":             "photo",
			"url":              "http://i.imgur.com/UmpOi.gif",
			"version":          "1.0",
			"width":            200,
		},
	},
}

func init() {
	listener.Register("^invite$", func(message *listener.BotMessage) bool {
		inviteURL, err := message.Sender.GetInvite()
		if err != nil {
			log.Printf("WTF: %s", err)
		}

		message.Reply(fmt.Sprintf("Here's your invite: %s", inviteURL))

		return false
	})

	listener.Register("^!mindblown$", func(message *listener.BotMessage) bool {
		body := map[string]interface{}{
			"text":        "Mind. Blown. http://i.imgur.com/UmpOi.gif",
			"annotations": annotation,
		}

		message.ReplyJSON(body)

		return false
	})
}
