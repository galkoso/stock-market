package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"stock-market/backend/internal/auth"
	notificationshub "stock-market/backend/internal/notifications"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/services"
	"stock-market/backend/internal/telegram"

	"github.com/gin-gonic/gin"
)

type NotificationsHandler struct {
	notifications *services.NotificationsService
	hub           *notificationshub.Hub
	telegram      *telegram.TelegramNotifier
}

func NewNotificationsHandler(
	notifications *services.NotificationsService,
	hub *notificationshub.Hub,
	telegramNotifier *telegram.TelegramNotifier,
) *NotificationsHandler {
	return &NotificationsHandler{
		notifications: notifications,
		hub:           hub,
		telegram:      telegramNotifier,
	}
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

func (h *NotificationsHandler) Stream(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "SSE_UNSUPPORTED", Message: "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	unread, _ := h.notifications.UnreadCount(c.Request.Context(), authUser.UserID)
	connectedPayload, _ := json.Marshal(gin.H{"unreadCount": unread})
	_, _ = c.Writer.Write(notificationshub.FormatSSE("connected", string(connectedPayload)))
	flusher.Flush()

	ch := h.hub.Subscribe(authUser.UserID)
	defer h.hub.Unsubscribe(authUser.UserID, ch)

	pingTicker := time.NewTicker(45 * time.Second)
	defer pingTicker.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			_, _ = fmt.Fprintf(c.Writer, ": keepalive\n\n")
			flusher.Flush()
		case msg, open := <-ch:
			if !open {
				return
			}
			_, _ = c.Writer.Write(msg)
			flusher.Flush()
		}
	}
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

type telegramTestRequest struct {
	Message string `json:"message"`
}

func (h *NotificationsHandler) TestTelegram(c *gin.Context) {
	if h.telegram == nil || !h.telegram.Enabled() {
		c.JSON(http.StatusServiceUnavailable, model.APIError{
			Code:    "TELEGRAM_NOT_CONFIGURED",
			Message: "Telegram bot is not configured. Set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID.",
		})
		return
	}

	var body telegramTestRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Message) == "" {
		c.JSON(http.StatusBadRequest, model.APIError{
			Code:    "INVALID_REQUEST",
			Message: "message is required",
		})
		return
	}

	if err := h.telegram.SendMessage(c.Request.Context(), body.Message); err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{
			Code:    "TELEGRAM_SEND_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
