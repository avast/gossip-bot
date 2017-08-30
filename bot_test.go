package main

import "testing"

func TestEvaluate(t *testing.T) {
	var message = mesg{
		messageID:    "1",
		messageText:  "An interesting message",
		channelID:    "1",
		channelName:  "A channel",
		userID:       "1",
		username:     "username",
		userRealname: "Name Surname",
		emojiCount:   2,
		replyCount:   2,
		timestamp:    "timestamp",
	}

	if !isMessageInteresting(message) {
		t.Fatalf("interesting message not marked as interesting")
	}
}
