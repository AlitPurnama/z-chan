package event

import (
	"log"
	"strings"

	"github.com/disgoorg/disgo/events"

	"discord-bot/internal/commands"
	"discord-bot/internal/utils"
)

func MessageCreate(e *events.MessageCreate) {
	if e.Message.Author.Bot {
		return
	}

	prefix, err := utils.GetPrefix(*e.GuildID)
	if err != nil {
		log.Printf("Error getting prefix: %v", err)
		return
	}
	msgContent := e.Message.Content
	command, _ := strings.CutPrefix(msgContent, prefix)

	command = strings.TrimSpace(command)

	parts := strings.Fields(command)

	if len(parts) > 0 {
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "ping":
			commands.HandlePingCommand(e, args...)
		case "prefix":
			commands.HandlePrefixCommand(e, args...)
		}

	}
}
