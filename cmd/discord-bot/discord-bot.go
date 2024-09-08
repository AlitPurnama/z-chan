package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-bot/config"
	"discord-bot/internal/services/cache"
	"discord-bot/internal/services/database"
	"discord-bot/internal/services/discord"
)

func main() {
	// Initialize configuration
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Error loading env: %v", err)
	}

	// Initialize database
	database.InitDatabase()
	defer database.CloseDatabase()

	// Initialize redis
	cache.InitRedis()
	defer cache.CloseRedis()

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	// Create and start the Discord bot
	if err := discord.CreateDiscordBot(ctx); err != nil {
		log.Fatalf("Error starting Discord bot: %v", err)
	}

	defer discord.CloseDiscordBot(ctx)

	// Wait for the context to be canceled (e.g., by receiving a signal)
	<-ctx.Done()

	log.Println("Service stopping...")
}
