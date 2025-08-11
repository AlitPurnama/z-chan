package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"discord-bot/internal/services/cache"
	"discord-bot/internal/services/database"
)

type GuildSettings struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	GuildID snowflake.ID       `bson:"guild_id"`
	Prefix  string             `bson:"prefix"`
}

func CreateGuildSettings(guildID *snowflake.ID) (*GuildSettings, error) {
	// GetGuildSettings handles cache retrieval and database lookup
	return GetGuildSettings(guildID)
}

func GetGuildSettings(guildID *snowflake.ID) (*GuildSettings, error) {
	// Get Guild Settings From Cache first
	settings, err := getGuildSettingsFromCache(*guildID)
	if err == nil && settings != nil {
		return settings, nil
	}

	ctx := context.Background()
	collection, err := database.GetCollection("guild_settings")
	if err != nil {
		return nil, fmt.Errorf("error getting MongoDB collection: %w", err)
	}

	var foundSettings GuildSettings
	filter := bson.M{"guild_id": guildID}

	err = collection.FindOne(ctx, filter).Decode(&foundSettings)
	if err == mongo.ErrNoDocuments {
		// No document found, create new settings
		settings = &GuildSettings{
			Prefix:  ">>",
			GuildID: *guildID,
		}

		_, err = collection.InsertOne(ctx, settings)
		if err != nil {
			return nil, fmt.Errorf("error inserting new guild settings: %w", err)
		}

		err = setGuildSettingsToCache(*settings)
		if err != nil {
			log.Printf("Warning: failed to cache new guild settings: %v", err)
		}
		return settings, nil
	} else if err != nil {
		// Other errors
		return nil, fmt.Errorf("error finding guild settings: %w", err)
	}

	// Successfully found and decoded the settings
	err = setGuildSettingsToCache(foundSettings)
	if err != nil {
		log.Printf("Warning: failed to cache guild settings: %v", err)
	}

	return &foundSettings, nil
}

func setGuildSettingsToCache(settings GuildSettings) error {
	rdb := cache.Client
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = rdb.Set(ctx, "guild_settings:"+settings.GuildID.String(), jsonData, 1*time.Minute).Err()
	log.Println("Set Cache")
	return err
}

func getGuildSettingsFromCache(guildID snowflake.ID) (*GuildSettings, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	val, err := cache.Client.Get(ctx, "guild_settings:"+guildID.String()).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("No guild settings found in cache for GID: %v", guildID)
	} else if err != nil {
		return nil, err
	}

	var settings GuildSettings
	if err := json.Unmarshal([]byte(val), &settings); err != nil {
		return nil, fmt.Errorf("Error deserializing guild settings for GID: %v", guildID)
	}

	log.Println("Get Cache")

	return &settings, nil
}

func GetPrefix(guildID snowflake.ID) (string, error) {
	settings, err := GetGuildSettings(&guildID)
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}
	return settings.Prefix, nil
}

func ChangePrefix(guildID snowflake.ID, newPrefix string) error {
	currentPrefix, err := GetPrefix(guildID)
	if err != nil {
		return err
	}

	if currentPrefix == newPrefix {
		return fmt.Errorf("warning: prefix is already set to %v", newPrefix)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collection, err := database.GetCollection("guild_settings")
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	filter := bson.M{
		"guild_id": guildID,
	}

	update := bson.M{
		"$set": bson.M{
			"prefix": newPrefix,
		},
	}
	var result GuildSettings
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&result)
	if err != nil {
		return fmt.Errorf("error updating document: %v", err)
	}
	setGuildSettingsToCache(result)
	return nil
}
