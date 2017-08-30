package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

type mesg struct {
	messageID    string
	messageText  string
	channelID    string
	channelName  string
	userID       string
	username     string
	userRealname string
	emojiCount   int
	replyCount   int
	timestamp    string
}

func main() {
	api := slack.New(os.Getenv("GOSSIPBOT_TOKEN"))
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	info, _ := api.GetTeamInfo()
	apiDomain := fmt.Sprintf("https://%s.slack.com", info.Domain)
	archivesRootUrl := fmt.Sprintf("%s/archives", apiDomain)

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
						messageText:  ev.Text,
						replyCount:   0,
						emojiCount:   0,
						channelID:    ev.Channel,
						channelName:  channel.Name,
						userID:       ev.User,
						username:     user.Name,
						userRealname: user.RealName,
						timestamp:    ev.Timestamp,
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

			var m = messages[ev.Item.Timestamp]
			if ev.Item.Timestamp != "" {
				m.emojiCount = m.emojiCount + 1
				messages[ev.Item.Timestamp] = m
			}

			if isMessageInteresting(m) {
				forwardMessage(m, rtm, archivesRootUrl)
			}

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

func isMessageInteresting(message mesg) bool {
	return message.emojiCount > 1
}

func forwardMessage(message mesg, rtm *slack.RTM, archivesUrl string) {
	messageToForward := rtm.NewOutgoingMessage(
		fmt.Sprintf("There is new interesting message %s", 
			fmt.Sprintf("%s/%s/p%s", 
				archivesUrl, 
				message.channelName, 
				strings.Replace(message.timestamp, ".", "", -1))),
		os.Getenv("GOSSIPBOT_CHANNEL"),
	)

	rtm.SendMessage(messageToForward)
}
