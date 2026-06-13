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

const (
	usersCollectionName         = "users"
	watchlistCollectionName     = "watchlist_items"
	alertsCollectionName        = "alerts"
	notificationsCollectionName = "notifications"
	portfolioCollectionName     = "portfolio_holdings"
)

type Database struct {
	client        *mongo.Client
	db            *mongo.Database
	Users         *mongo.Collection
	Watchlist     *mongo.Collection
	Alerts        *mongo.Collection
	Notifications *mongo.Collection
	Portfolio     *mongo.Collection
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
	watchlist := db.Collection(watchlistCollectionName)
	alerts := db.Collection(alertsCollectionName)
	notifications := db.Collection(notificationsCollectionName)
	portfolio := db.Collection(portfolioCollectionName)

	indexCtx, indexCancel := context.WithTimeout(ctx, 15*time.Second)
	defer indexCancel()

	if err := ensureIndexes(indexCtx, users, watchlist, alerts, notifications, portfolio); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return &Database{
		client:        client,
		db:            db,
		Users:         users,
		Watchlist:     watchlist,
		Alerts:        alerts,
		Notifications: notifications,
		Portfolio:     portfolio,
	}, nil
}

func ensureIndexes(ctx context.Context, users, watchlist, alerts, notifications, portfolio *mongo.Collection) error {
	if _, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("create users index: %w", err)
	}

	if _, err := watchlist.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "symbol", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("create watchlist index: %w", err)
	}

	if _, err := watchlist.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create watchlist user index: %w", err)
	}

	if _, err := alerts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create alerts user index: %w", err)
	}

	if _, err := alerts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "isActive", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create alerts active index: %w", err)
	}

	if _, err := notifications.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create notifications user index: %w", err)
	}

	if _, err := notifications.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "isRead", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create notifications unread index: %w", err)
	}

	if _, err := notifications.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "alertId", Value: 1}, {Key: "createdAt", Value: -1}},
	}); err != nil {
		return fmt.Errorf("create notifications alert index: %w", err)
	}

	if _, err := portfolio.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "symbol", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("create portfolio index: %w", err)
	}

	if _, err := portfolio.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	}); err != nil {
		return fmt.Errorf("create portfolio user index: %w", err)
	}

	return nil
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
