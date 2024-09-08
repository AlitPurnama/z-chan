package discord

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"

	"discord-bot/internal/event"
)

var client bot.Client

func CreateDiscordBot(ctx context.Context) error {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("can't find bot token")
	}

	var err error
	client, err = disgo.New(botToken, bot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentsAll,
		),
	))
	if err != nil {
		return err
	}

	if err = client.OpenGateway(ctx); err != nil {
		return err
	}

	registerEvents(client)

	return nil
}

func CloseDiscordBot(ctx context.Context) {
	if client != nil {
		client.Close(ctx)
		log.Println("Discord bot stopped")
	}
}

func registerEvents(c bot.Client) {
	log.Println("Registering events...")
	c.AddEventListeners(&events.ListenerAdapter{
		OnReady:         event.EventReady,
		OnMessageCreate: event.MessageCreate,
	})
}
