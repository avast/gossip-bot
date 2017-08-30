package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

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

func (message mesg) isMessageImportant() bool {
	return message.emojiCount > 1
}

func main() {
	api := slack.New(os.Getenv("GOSSIPBOT_TOKEN"))

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	info, err := api.GetTeamInfo()
	if err != nil {
		log.Fatal(err)
	}
	archivesRootURL := fmt.Sprintf("https://%s.slack.com/archives", info.Domain)

	messages := make(map[string]mesg, 0)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.ConnectedEvent:
			log.Debug("Connection counter:", ev.ConnectionCount)

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

				log.Debug("== Processing new message event: \n")
				log.Debug("message object: %+v\n", ev)
				log.Debug("Message: %s\n", ev.Text)
				log.Debug("Subtype: %s\n", ev.SubType)
				log.Debug("Timestamp: %s\n", ev.Timestamp)
				log.Debug("ThreadTimestamp: %s\n", ev.ThreadTimestamp)
				log.Debug("Messages: %+v\n", messages)
			}

		case *slack.ReactionAddedEvent:
			log.Debug("-------------\n")
			log.Debug("Reaction: %+v\n", ev)

			var m = messages[ev.Item.Timestamp]
			if ev.Item.Timestamp != "" {
				m.emojiCount = m.emojiCount + 1
				messages[ev.Item.Timestamp] = m
			}

			if m.isMessageImportant() {
				forwardMessage(m, rtm, archivesRootURL)
			}

		case *slack.ReactionRemovedEvent:
			log.Debug("-------------\n")
			log.Debug("Reaction removed: %+v\n", ev)

			if ev.Item.Timestamp != "" {
				var m = messages[ev.Item.Timestamp]
				m.emojiCount = m.emojiCount - 1
				messages[ev.Item.Timestamp] = m
			}

		case *slack.InvalidAuthEvent:
			log.Error("Invalid credentials")
			return

		default:
		}
	}
}

func forwardMessage(message mesg, rtm *slack.RTM, archivesURL string) {
	messageToForward := rtm.NewOutgoingMessage(
		fmt.Sprintf("There is new interesting message %s/%s/p%s",
			archivesURL,
			message.channelName,
			strings.Replace(message.timestamp, ".", "", -1)),
		os.Getenv("GOSSIPBOT_CHANNEL"),
	)

	rtm.SendMessage(messageToForward)
}
