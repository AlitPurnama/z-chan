package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/events"
)

func HandlePingCommand(e *events.MessageCreate, args ...string) {
	fmt.Println(e.Message.Content, args)
}
