package main

import (
	"fmt"
	"log"
	"mxml/makini/api"
	"mxml/makini/listener"
	"strings"
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

func formatRemainingCount(remainingCount int) string {
	if remainingCount < 1 {
		return "You have no more invites available at the moment."
	} else if remainingCount == 1 {
		return "You've got one invite left."
	}

	return fmt.Sprintf("You have %d invites available.", remainingCount)
}

func init() {
	listener.Register("(?i)^(?:what'?s|get )?(?:my )?invite(?: count)?\\??$", func(message *listener.BotMessage) bool {
		remainingCount, err := message.Sender.GetInviteCount()
		if err != nil {
			log.Printf("Error getting invite count for %s: %s", message.Sender.Username(), err)
			message.Reply("Sorry, I had trouble getting your invite count. Please try later.")
		}

		message.Reply(formatRemainingCount(remainingCount))

		return true
	})

	listener.Register("(?i)^(?:get|give) (?:me )?(?:an )?invite(?: .+)?$", func(message *listener.BotMessage) bool {
		inviteURL, remainingCount, err := message.Sender.GetInvite("")
		if err != nil {
			log.Printf("Error getting invite for %s: %s", message.Sender.Username(), err)

			if meta, ok := err.(*api.APIMeta); ok && meta.ErrorSlug == "invites_depleted" {
				message.Reply("Sorry, it looks like you don't have any invites available.")

				return true
			}

			message.Reply("Sorry, I couldn't get an invite for you.")

			return true
		}

		message.Reply(fmt.Sprintf("Here's the link for your invite: %s. %s", inviteURL, formatRemainingCount(remainingCount)))

		return true
	})

	listener.Register("(?i)^send (?:an )?invite (?:to )?(\\S+)?$", func(message *listener.BotMessage) bool {
		email := message.Matches[1]
		if !strings.Contains(email, "@") {
			log.Printf("Invalid email from %s: %s", message.Sender.Username(), email)
			message.Reply(fmt.Sprintf("Hmm. %s doesn't look like an email address to me.", email))

			return true
		}

		_, remainingCount, err := message.Sender.GetInvite(email)
		if err != nil {
			log.Printf("Error getting invite for %s: %s", message.Sender.Username(), err)

			if meta, ok := err.(*api.APIMeta); ok && meta.ErrorSlug == "invites_depleted" {
				message.Reply("Sorry, it looks like you don't have any invites available.")

				return true
			}

			message.Reply("Sorry, I couldn't get an invite for you.")

			return true
		}

		message.Reply(fmt.Sprintf("OK, I sent an email to %s. %s", email, formatRemainingCount(remainingCount)))

		return true
	})

	listener.Register("(?i)^!mindblown$", func(message *listener.BotMessage) bool {
		body := map[string]interface{}{
			"text":        "Mind. Blown. http://i.imgur.com/UmpOi.gif",
			"annotations": annotation,
		}

		message.ReplyJSON(body)

		return true
	})

	listener.Register("(?i)^help computer$", func(message *listener.BotMessage) bool {
		body := map[string]interface{}{
			"text": "I'm a computer. Stop all the downloadin'. Err... I mean, uh, that you should probably email my human friends at support@app.net for more help.",
		}

		message.ReplyJSON(body)

		return true
	})

	listener.Register("(?i)^help( .+)?$", func(message *listener.BotMessage) bool {
		body := map[string]interface{}{
			"text": "Try asking me to 'send invite to <email>' or 'get invite link' or ask me 'invite count?' and I'll try to help. For more complex questions, email my human friends at support@app.net.",
		}

		message.ReplyJSON(body)

		return true
	})

	listener.Register(".", func(message *listener.BotMessage) bool {
		if !message.Sender.Flags.SentIntro {
			body := map[string]interface{}{
				"text": "Hi, I'm @adn, a bot who's here to help you with your App.net account. Try asking me to 'send invite to <email>' or 'get invite link' or ask me 'invite count?' and I'll try to help.",
			}

			message.ReplyJSON(body)
			message.Sender.Flags.SentIntro = true
		}

		return true
	})
}
