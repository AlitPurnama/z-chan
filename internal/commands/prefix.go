package commands

import (
	"fmt"
	"log"

	"github.com/disgoorg/disgo/events"

	"discord-bot/internal/utils"
)

func HandlePrefixCommand(e *events.MessageCreate, args ...string) {
	argsLength := len(args)

	if argsLength == 0 {
		return
	}

	cmdType := args[0]
	if cmdType == "" {
		return
	}

	switch cmdType {
	case "change":
		if argsLength < 2 {
			msg := utils.CreateMessage("You need to specify your new prefix")
			utils.ReplyMessage(e, msg)
			return
		}
		newPrefix := args[1]
		err := utils.ChangePrefix(*e.GuildID, newPrefix)
		if err != nil {
			log.Printf("error changing prefix for: %v, error: %v", e.GuildID, err)
			msg := utils.CreateMessage("Error changing your prefix...")
			utils.ReplyMessage(e, msg)
			return
		}
		msg := utils.CreateMessage("Prefix has changed to: " + "`" + newPrefix + "`")
		utils.ReplyMessage(e, msg)
	case "set":

	default:
		fmt.Println("Default")
	}
}
