package utils

import (
	"context"
	"fmt"
	"testing"

	"github.com/disgoorg/snowflake/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockCollection struct {
	data map[string]GuildSettings
}

func newMockCollection() *mockCollection {
	return &mockCollection{data: make(map[string]GuildSettings)}
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	var gid string
	if f, ok := filter.(bson.M); ok {
		switch v := f["guild_id"].(type) {
		case *snowflake.ID:
			gid = v.String()
		case snowflake.ID:
			gid = v.String()
		}
	}
	if gs, ok := m.data[gid]; ok {
		return mongo.NewSingleResultFromDocument(gs, nil, nil)
	}
	return mongo.NewSingleResultFromDocument(bson.M{}, mongo.ErrNoDocuments, nil)
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	gs := document.(*GuildSettings)
	m.data[gs.GuildID.String()] = *gs
	return &mongo.InsertOneResult{InsertedID: gs.ID}, nil
}

func (m *mockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	var gid string
	if f, ok := filter.(bson.M); ok {
		switch v := f["guild_id"].(type) {
		case *snowflake.ID:
			gid = v.String()
		case snowflake.ID:
			gid = v.String()
		}
	}
	gs := m.data[gid]
	if u, ok := update.(bson.M); ok {
		if set, ok := u["$set"].(bson.M); ok {
			if p, ok := set["prefix"].(string); ok {
				gs.Prefix = p
				m.data[gid] = gs
			}
		}
	}
	return mongo.NewSingleResultFromDocument(gs, nil, nil)
}

// setupMocks overrides package level dependencies with in-memory versions.
func setupMocks() (*mockCollection, map[string]GuildSettings, func()) {
	mc := newMockCollection()
	cacheStore := make(map[string]GuildSettings)

	origGetCollection := getCollection
	origGetCache := getGuildSettingsFromCacheFn
	origSetCache := setGuildSettingsToCacheFn

	getCollection = func(name string) (guildCollection, error) { return mc, nil }
	getGuildSettingsFromCacheFn = func(id snowflake.ID) (*GuildSettings, error) {
		if gs, ok := cacheStore[id.String()]; ok {
			val := gs
			return &val, nil
		}
		return nil, fmt.Errorf("cache miss")
	}
	setGuildSettingsToCacheFn = func(gs GuildSettings) error {
		cacheStore[gs.GuildID.String()] = gs
		return nil
	}

	cleanup := func() {
		getCollection = origGetCollection
		getGuildSettingsFromCacheFn = origGetCache
		setGuildSettingsToCacheFn = origSetCache
	}

	return mc, cacheStore, cleanup
}

func TestChangePrefixAndCaching(t *testing.T) {
	mc, cacheStore, cleanup := setupMocks()
	defer cleanup()

	guildID := snowflake.ID(12345)

	// Initial prefix should be default and stored in both DB and cache.
	prefix, err := GetPrefix(guildID)
	if err != nil {
		t.Fatalf("GetPrefix error: %v", err)
	}
	if prefix != ">>" {
		t.Fatalf("expected default prefix '>>', got %s", prefix)
	}
	if mc.data[guildID.String()].Prefix != ">>" {
		t.Fatalf("database not initialized with default prefix")
	}
	if cacheStore[guildID.String()].Prefix != ">>" {
		t.Fatalf("cache not initialized with default prefix")
	}

	// Change prefix and ensure it updates DB, cache, and GetPrefix returns new value.
	if err := ChangePrefix(guildID, "!!"); err != nil {
		t.Fatalf("ChangePrefix error: %v", err)
	}
	if mc.data[guildID.String()].Prefix != "!!" {
		t.Fatalf("database not updated with new prefix")
	}
	if cacheStore[guildID.String()].Prefix != "!!" {
		t.Fatalf("cache not updated with new prefix")
	}
	prefix, err = GetPrefix(guildID)
	if err != nil {
		t.Fatalf("GetPrefix error: %v", err)
	}
	if prefix != "!!" {
		t.Fatalf("expected updated prefix '!!', got %s", prefix)
	}

	// Modify database and ensure cached value is returned.
	gs := mc.data[guildID.String()]
	gs.Prefix = "??"
	mc.data[guildID.String()] = gs
	prefix, err = GetPrefix(guildID)
	if err != nil {
		t.Fatalf("GetPrefix error: %v", err)
	}
	if prefix != "!!" {
		t.Fatalf("expected cached prefix '!!', got %s", prefix)
	}

	// Clear cache and ensure DB value is returned.
	delete(cacheStore, guildID.String())
	prefix, err = GetPrefix(guildID)
	if err != nil {
		t.Fatalf("GetPrefix error: %v", err)
	}
	if prefix != "??" {
		t.Fatalf("expected DB prefix '??' after cache clear, got %s", prefix)
	}
}
