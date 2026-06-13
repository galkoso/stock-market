package services

import (
	"context"
	"fmt"
	"strings"

	"stock-market/backend/internal/repositories"
)

type NotificationsService struct {
	repo *repositories.NotificationsRepository
}

func NewNotificationsService(repo *repositories.NotificationsRepository) *NotificationsService {
	return &NotificationsService{repo: repo}
}

func (s *NotificationsService) List(ctx context.Context, userID string) ([]repositories.Notification, error) {
	return s.repo.List(ctx, userID, 50)
}

func (s *NotificationsService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *NotificationsService) MarkRead(ctx context.Context, userID, notificationID string) error {
	if strings.TrimSpace(notificationID) == "" {
		return fmt.Errorf("notification id is required")
	}
	return s.repo.MarkRead(ctx, userID, notificationID)
}

func (s *NotificationsService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}
