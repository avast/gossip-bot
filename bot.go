package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nlopes/slack"
)

type mesg struct {
	text       string
	emojiCount int
	replyCount int
	timestamp  string
}

func main() {
	api := slack.New(os.Getenv("GOSSIPBOT_TOKEN"))
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	//api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	messages := make(map[string]mesg, 0)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.ConnectedEvent:
			//fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			if ev.SubType == "" {
				if ev.ThreadTimestamp != "" {
					// messages[ev.ThreadTimestamp].count++
					var m = messages[ev.ThreadTimestamp]
					m.replyCount = m.replyCount + 1
					messages[ev.ThreadTimestamp] = m
				} else {
					messages[ev.Timestamp] = mesg{
						text:       ev.Text,
						replyCount: 0,
						emojiCount: 0,
						timestamp:  ev.Timestamp,
					}
				}

				fmt.Printf("-------------\n")
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

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			//			fmt.Printf(".")
		}
	}
}
