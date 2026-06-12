package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Notification struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"userId"`
	AlertID   string    `json:"alertId,omitempty" bson:"alertId,omitempty"`
	Symbol    string    `json:"symbol,omitempty" bson:"symbol,omitempty"`
	Title     string    `json:"title" bson:"title"`
	Message   string    `json:"message" bson:"message"`
	IsRead    bool      `json:"isRead" bson:"isRead"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type NotificationsRepository struct {
	collection *mongo.Collection
}

func NewNotificationsRepository(collection *mongo.Collection) *NotificationsRepository {
	return &NotificationsRepository{collection: collection}
}

func (r *NotificationsRepository) Create(ctx context.Context, userID, alertID, symbol, title, message string) (*Notification, error) {
	notification := Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		AlertID:   alertID,
		Symbol:    symbol,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if _, err := r.collection.InsertOne(ctx, notification); err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationsRepository) List(ctx context.Context, userID string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	cursor, err := r.collection.Find(
		ctx,
		bson.M{"userId": userID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]Notification, 0)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *NotificationsRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"userId": userID, "isRead": false})
	return int(count), err
}

func (r *NotificationsRepository) MarkRead(ctx context.Context, userID, notificationID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": notificationID, "userId": userID},
		bson.M{"$set": bson.M{"isRead": true}},
	)
	return err
}

func (r *NotificationsRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userID, "isRead": false},
		bson.M{"$set": bson.M{"isRead": true}},
	)
	return err
}

func (r *NotificationsRepository) ExistsRecentForAlert(ctx context.Context, alertID string, since time.Time) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"alertId":   alertID,
		"createdAt": bson.M{"$gte": since},
	})
	return count > 0, err
}
