package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Alert struct {
	ID              string         `json:"id" bson:"_id"`
	UserID          string         `json:"userId" bson:"userId"`
	Symbol          string         `json:"symbol,omitempty" bson:"symbol,omitempty"`
	AlertType       string         `json:"alertType" bson:"alertType"`
	Params          map[string]any `json:"params" bson:"params"`
	IsActive        bool           `json:"isActive" bson:"isActive"`
	LastTriggeredAt *time.Time     `json:"lastTriggeredAt,omitempty" bson:"lastTriggeredAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt" bson:"createdAt"`
}

type AlertsRepository struct {
	collection *mongo.Collection
}

func NewAlertsRepository(collection *mongo.Collection) *AlertsRepository {
	return &AlertsRepository{collection: collection}
}

func (r *AlertsRepository) List(ctx context.Context, userID string) ([]Alert, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	alerts := make([]Alert, 0)
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	for i := range alerts {
		if alerts[i].Params == nil {
			alerts[i].Params = map[string]any{}
		}
	}

	return alerts, nil
}

func (r *AlertsRepository) Create(ctx context.Context, userID, symbol, alertType string, params map[string]any) (*Alert, error) {
	if params == nil {
		params = map[string]any{}
	}

	alert := Alert{
		ID:        uuid.NewString(),
		UserID:    userID,
		Symbol:    symbol,
		AlertType: alertType,
		Params:    params,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if _, err := r.collection.InsertOne(ctx, alert); err != nil {
		return nil, err
	}

	return &alert, nil
}

func (r *AlertsRepository) Delete(ctx context.Context, userID, alertID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": alertID, "userId": userID})
	return err
}

func (r *AlertsRepository) MarkTriggered(ctx context.Context, alertID string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": alertID}, bson.M{"$set": bson.M{"lastTriggeredAt": now}})
	return err
}

func (r *AlertsRepository) ListActive(ctx context.Context) ([]Alert, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"isActive": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	alerts := make([]Alert, 0)
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	for i := range alerts {
		if alerts[i].Params == nil {
			alerts[i].Params = map[string]any{}
		}
	}

	return alerts, nil
}
