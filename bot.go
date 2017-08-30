package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nlopes/slack"
)

type mesg struct {
	message_id    string
	message_text  string
	channel_id    string
	channel_name  string
	user_id       string
	user_name     string
	user_realname string
	emojiCount    int
	replyCount    int
	timestamp     string
}

func main() {
	api := slack.New(os.Getenv("GOSSIPBOT_TOKEN"))
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	messages := make(map[string]mesg, 0)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.ConnectedEvent:
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			if ev.SubType == "" {
				if ev.ThreadTimestamp != "" {
					var m = messages[ev.ThreadTimestamp]
					m.replyCount = m.replyCount + 1
					messages[ev.ThreadTimestamp] = m
				} else {
					channel, _ := api.GetChannelInfo(ev.Channel)
					user, _ := api.GetUserInfo(ev.User)

					messages[ev.Timestamp] = mesg{
						message_text:  ev.Text,
						replyCount:    0,
						emojiCount:    0,
						channel_id:    ev.Channel,
						channel_name:  channel.Name,
						user_id:       ev.User,
						user_name:     user.Name,
						user_realname: user.RealName,
						timestamp:     ev.Timestamp,
					}
				}

				fmt.Printf("-------------\n")
				fmt.Printf("message object: %+v\n", ev)
				fmt.Printf("Message: %s\n", ev.Text)
				fmt.Printf("Subtype: %s\n", ev.SubType)
				fmt.Printf("Timestamp: %s\n", ev.Timestamp)
				fmt.Printf("ThreadTimestamp: %s\n", ev.ThreadTimestamp)
				fmt.Printf("Messages: %+v\n", messages)
			}

		case *slack.ReactionAddedEvent:
			fmt.Printf("-------------\n")
			fmt.Printf("Reaction: %+v\n", ev)

			if ev.Item.Timestamp != "" {
				var m = messages[ev.Item.Timestamp]
				m.emojiCount = m.emojiCount + 1
				messages[ev.Item.Timestamp] = m
			}

			evaluateMessage(messages[ev.Item.Timestamp], rtm)
		case *slack.ReactionRemovedEvent:
			fmt.Printf("-------------\n")
			fmt.Printf("Reaction removed: %+v\n", ev)

			if ev.Item.Timestamp != "" {
				var m = messages[ev.Item.Timestamp]
				m.emojiCount = m.emojiCount - 1
				messages[ev.Item.Timestamp] = m
			}

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			//			fmt.Printf(".")
		}
	}
}

func evaluateMessage(message mesg, rtm *slack.RTM) {
	if message.emojiCount > 1 {
		message_to_forward := rtm.NewOutgoingMessage(
			fmt.Sprintf("There is an interesting message '%s' from %s in <#%s|%s>",
				message.message_text,
				message.user_realname,
				message.channel_id,
				message.channel_name),
			os.Getenv("GOSSIPBOT_CHANNEL"))

		rtm.SendMessage(message_to_forward)
	}
}
