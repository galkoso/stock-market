package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WatchlistItem struct {
	ID          string    `json:"id" bson:"_id"`
	UserID      string    `json:"userId" bson:"userId"`
	Symbol      string    `json:"symbol" bson:"symbol"`
	CompanyName string    `json:"companyName" bson:"companyName"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type WatchlistRepository struct {
	collection *mongo.Collection
}

func NewWatchlistRepository(collection *mongo.Collection) *WatchlistRepository {
	return &WatchlistRepository{collection: collection}
}

func (r *WatchlistRepository) List(ctx context.Context, userID string) ([]WatchlistItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]WatchlistItem, 0)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *WatchlistRepository) Add(ctx context.Context, userID, symbol, companyName string) (*WatchlistItem, error) {
	now := time.Now()
	filter := bson.M{"userId": userID, "symbol": symbol}
	update := bson.M{
		"$set": bson.M{"companyName": companyName},
		"$setOnInsert": bson.M{
			"_id":       uuid.NewString(),
			"userId":    userID,
			"symbol":    symbol,
			"createdAt": now,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var item WatchlistItem
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *WatchlistRepository) Remove(ctx context.Context, userID, symbol string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"userId": userID, "symbol": symbol})
	return err
}

func (r *WatchlistRepository) Symbols(ctx context.Context, userID string) ([]string, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	symbols := make([]string, 0)
	for cursor.Next(ctx) {
		var item WatchlistItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		symbols = append(symbols, item.Symbol)
	}

	return symbols, cursor.Err()
}
