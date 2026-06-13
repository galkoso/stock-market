package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PortfolioHolding struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"userId"`
	Symbol    string    `json:"symbol" bson:"symbol"`
	Quantity  float64   `json:"quantity" bson:"quantity"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type PortfolioRepository struct {
	collection *mongo.Collection
}

func NewPortfolioRepository(collection *mongo.Collection) *PortfolioRepository {
	return &PortfolioRepository{collection: collection}
}

func (r *PortfolioRepository) List(ctx context.Context, userID string) ([]PortfolioHolding, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	holdings := make([]PortfolioHolding, 0)
	if err := cursor.All(ctx, &holdings); err != nil {
		return nil, err
	}

	return holdings, nil
}

func (r *PortfolioRepository) FindBySymbol(ctx context.Context, userID, symbol string) (*PortfolioHolding, error) {
	var holding PortfolioHolding
	err := r.collection.FindOne(ctx, bson.M{"userId": userID, "symbol": symbol}).Decode(&holding)
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

func (r *PortfolioRepository) Create(ctx context.Context, userID, symbol string, quantity float64) (*PortfolioHolding, error) {
	now := time.Now()
	holding := PortfolioHolding{
		ID:        uuid.NewString(),
		UserID:    userID,
		Symbol:    symbol,
		Quantity:  quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := r.collection.InsertOne(ctx, holding); err != nil {
		return nil, err
	}

	return &holding, nil
}

func (r *PortfolioRepository) UpdateQuantity(ctx context.Context, userID, symbol string, quantity float64) (*PortfolioHolding, error) {
	now := time.Now()
	filter := bson.M{"userId": userID, "symbol": symbol}
	update := bson.M{
		"$set": bson.M{
			"quantity":  quantity,
			"updatedAt": now,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var holding PortfolioHolding
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&holding); err != nil {
		return nil, err
	}

	return &holding, nil
}

func (r *PortfolioRepository) Remove(ctx context.Context, userID, symbol string) (int64, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{"userId": userID, "symbol": symbol})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
