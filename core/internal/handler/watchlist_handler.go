package handler

import (
	"net/http"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type watchlistAddRequest struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"companyName"`
}

type WatchlistHandler struct {
	watchlist *services.WatchlistService
}

func NewWatchlistHandler(watchlist *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlist: watchlist}
}

func (h *WatchlistHandler) List(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	items, err := h.watchlist.List(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "WATCHLIST_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *WatchlistHandler) Add(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	var body watchlistAddRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "INVALID_REQUEST", Message: "symbol is required"})
		return
	}

	item, err := h.watchlist.Add(c.Request.Context(), authUser.UserID, body.Symbol, body.CompanyName)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "WATCHLIST_ADD_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *WatchlistHandler) Remove(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.watchlist.Remove(c.Request.Context(), authUser.UserID, c.Param("symbol")); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "WATCHLIST_REMOVE_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
