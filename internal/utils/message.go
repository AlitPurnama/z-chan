package utils

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func CreateMessage(content string) discord.MessageCreate {
	msg := discord.NewMessageCreateBuilder()
	msg.SetContent(content)
	return msg.Build()
}

func ReplyMessage(e *events.MessageCreate, message discord.MessageCreate) {
	if message.Content != "" {
		e.Client().Rest().CreateMessage(e.ChannelID, message)
	}
}
