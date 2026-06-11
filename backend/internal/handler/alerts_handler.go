package handler

import (
	"net/http"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AlertsHandler struct {
	alerts      *services.AlertsService
	alertEngine *services.AlertEngine
}

func NewAlertsHandler(alerts *services.AlertsService, alertEngine *services.AlertEngine) *AlertsHandler {
	return &AlertsHandler{alerts: alerts, alertEngine: alertEngine}
}

type createAlertRequest struct {
	Symbol    string         `json:"symbol"`
	AlertType string         `json:"alertType"`
	Params    map[string]any `json:"params"`
}

func (h *AlertsHandler) List(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	items, err := h.alerts.List(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "ALERTS_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alerts": items})
}

func (h *AlertsHandler) Create(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	var body createAlertRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.AlertType == "" {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "INVALID_REQUEST", Message: "alertType is required"})
		return
	}

	alert, err := h.alerts.Create(c.Request.Context(), authUser.UserID, body.Symbol, body.AlertType, body.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "ALERT_CREATE_FAILED", Message: err.Error()})
		return
	}

	h.alertEngine.Evaluate(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{"alert": alert})
}

func (h *AlertsHandler) Evaluate(c *gin.Context) {
	if _, ok := auth.GetAuthUser(c); !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	h.alertEngine.Evaluate(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AlertsHandler) Delete(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.alerts.Delete(c.Request.Context(), authUser.UserID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "ALERT_DELETE_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
