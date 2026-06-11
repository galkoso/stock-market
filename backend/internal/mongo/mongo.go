package mongo

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"stock-market/backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const usersCollectionName = "users"

type Database struct {
	client *mongo.Client
	db     *mongo.Database
	Users  *mongo.Collection
}

func Connect(ctx context.Context, connectionString string) (*Database, error) {
	databaseName, err := databaseNameFromURI(connectionString)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(databaseName)

	users := db.Collection(usersCollectionName)
	indexCtx, indexCancel := context.WithTimeout(ctx, 10*time.Second)
	defer indexCancel()

	_, err = users.Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("create users index: %w", err)
	}

	return &Database{
		client: client,
		db:     db,
		Users:  users,
	}, nil
}

func (d *Database) Close(ctx context.Context) error {
	if d.client == nil {
		return nil
	}
	return d.client.Disconnect(ctx)
}

func databaseNameFromURI(connectionString string) (string, error) {
	parsed, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("parse mongo connection string: %w", err)
	}

	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if idx := strings.Index(databaseName, "?"); idx >= 0 {
		databaseName = databaseName[:idx]
	}

	if databaseName == "" {
		return "", fmt.Errorf("mongo connection string must include a database name")
	}

	return databaseName, nil
}

func ToSafeUser(user model.User) model.SafeUser {
	return model.SafeUser{
		ID:       user.ID,
		Username: user.Username,
	}
}
