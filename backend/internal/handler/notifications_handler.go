package handler

import (
	"net/http"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type NotificationsHandler struct {
	notifications *services.NotificationsService
}

func NewNotificationsHandler(notifications *services.NotificationsService) *NotificationsHandler {
	return &NotificationsHandler{notifications: notifications}
}

func (h *NotificationsHandler) List(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	items, err := h.notifications.List(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "NOTIFICATIONS_FAILED", Message: err.Error()})
		return
	}

	unread, err := h.notifications.UnreadCount(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "NOTIFICATIONS_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": items, "unreadCount": unread})
}

func (h *NotificationsHandler) MarkRead(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.notifications.MarkRead(c.Request.Context(), authUser.UserID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "NOTIFICATION_READ_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *NotificationsHandler) MarkAllRead(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.notifications.MarkAllRead(c.Request.Context(), authUser.UserID); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "NOTIFICATION_READ_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
