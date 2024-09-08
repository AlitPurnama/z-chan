package event

import (
	"log"

	"github.com/disgoorg/disgo/events"
)

func EventReady(e *events.Ready) {
	log.Printf("%v is ready", e.User.Username)
}
