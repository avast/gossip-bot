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

var messages map[string]mesg

func main() {
	api := slack.New(os.Getenv("GOSSIPBOT_BOT_TOKEN"))
	api_client := slack.New(os.Getenv("GOSSIPBOT_CLIENT_TOKEN"))

	log.SetLevel(log.DebugLevel)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// var search, error2, error3 = api_client.Search("1504103389.000400", slack.SearchParameters{Count: 1})
	// log.Debug(fmt.Sprintf("%+v", search))
	// log.Debug(fmt.Sprintf("%+v", error2))
	// log.Debug(fmt.Sprintf("%+v", error3))
	//
	var history, error1 = api_client.GetChannelHistory(os.Getenv("GOSSIPBOT_CHANNEL"), slack.HistoryParameters{Count: 1})
	log.Debug(fmt.Sprintf("%+v", history))
	log.Debug(fmt.Sprintf("%+v", error1))

	info, err := api.GetTeamInfo()
	if err != nil {
		log.Fatal(err)
	}
	archivesRootURL := fmt.Sprintf("https://%s.slack.com/archives", info.Domain)

	messages = make(map[string]mesg)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.ConnectedEvent:
			log.Debug("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			processMessageEvent(api, ev)

		case *slack.ReactionAddedEvent:
			processReactionAddedEvent(api, ev)
			if m := messages[ev.EventTimestamp]; m.isMessageImportant() {
				forwardMessage(m, rtm, archivesRootURL)
			}

		case *slack.ReactionRemovedEvent:
			processReactionRemovedEvent(api, ev)

		case *slack.InvalidAuthEvent:
			log.Error("Invalid credentials")
			return

		default:
		}
	}
}

func processMessageEvent(api *slack.Client, ev *slack.MessageEvent) {
	if ev.SubType == "" {
		if ev.ThreadTimestamp != "" {
			var m = messages[ev.ThreadTimestamp]
			m.replyCount = m.replyCount + 1
			messages[ev.ThreadTimestamp] = m
		} else {
			channel, err := api.GetChannelInfo(ev.Channel)
			if err != nil {
				log.Fatal(err)
			}
			user, err := api.GetUserInfo(ev.User)
			if err != nil {
				log.Fatal(err)
			}

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
}

func processReactionAddedEvent(api *slack.Client, ev *slack.ReactionAddedEvent) {
	log.Debug("-------------\n")
	log.Debug("Reaction: %+v\n", ev)

	m, present := messages[ev.Item.Timestamp]
	if !present {
		foundMessages, err := api.GetChannelHistory(ev.Item.Channel, slack.HistoryParameters{Latest: ev.Item.Timestamp, Inclusive: true, Count: 1})
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if len(foundMessages.Messages) > 0 {
			foundMessage := foundMessages.Messages[0]
			user, err := api.GetUserInfo(foundMessage.User)
			if err != nil {
				log.Fatalf("%+v", err)
			}
			channel, err := api.GetChannelInfo(ev.Item.Channel)
			if err != nil {
				log.Fatalf("%+v", err)
			}

			messages[ev.Item.Timestamp] = mesg{
				messageText:  foundMessage.Text,
				replyCount:   foundMessage.ReplyCount,
				emojiCount:   len(foundMessage.Reactions),
				channelID:    ev.Item.Channel,
				channelName:  channel.Name,
				userID:       foundMessage.User,
				username:     user.Name,
				userRealname: user.RealName,
				timestamp:    ev.Item.Timestamp,
			}
		} else {
			fmt.Printf("Message not found when calling API with timestamp %s", ev.Item.Timestamp)
		}
	}
	if ev.Item.Timestamp != "" {
		m.emojiCount = m.emojiCount + 1
		messages[ev.Item.Timestamp] = m
	}
}

func processReactionRemovedEvent(api *slack.Client, ev *slack.ReactionRemovedEvent) {
	log.Debug("-------------\n")
	log.Debug("Reaction removed: %+v\n", ev)

	if ev.Item.Timestamp != "" {
		var m = messages[ev.Item.Timestamp]
		m.emojiCount = m.emojiCount - 1
		messages[ev.Item.Timestamp] = m
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
