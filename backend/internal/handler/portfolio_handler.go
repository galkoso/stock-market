package handler

import (
	"net/http"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type portfolioAddRequest struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
}

type portfolioUpdateRequest struct {
	Quantity float64 `json:"quantity"`
}

type PortfolioHandler struct {
	portfolio *services.PortfolioService
}

func NewPortfolioHandler(portfolio *services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{portfolio: portfolio}
}

func (h *PortfolioHandler) List(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	holdings, err := h.portfolio.List(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "PORTFOLIO_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"holdings": holdings})
}

func (h *PortfolioHandler) Add(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	var body portfolioAddRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Symbol == "" {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "INVALID_REQUEST", Message: "symbol and quantity are required"})
		return
	}

	holding, err := h.portfolio.Add(c.Request.Context(), authUser.UserID, body.Symbol, body.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "PORTFOLIO_ADD_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"holding": holding})
}

func (h *PortfolioHandler) Update(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	var body portfolioUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "INVALID_REQUEST", Message: "quantity is required"})
		return
	}

	holding, err := h.portfolio.UpdateQuantity(c.Request.Context(), authUser.UserID, c.Param("symbol"), body.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "PORTFOLIO_UPDATE_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"holding": holding})
}

func (h *PortfolioHandler) Remove(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.portfolio.Remove(c.Request.Context(), authUser.UserID, c.Param("symbol")); err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "PORTFOLIO_REMOVE_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *PortfolioHandler) Allocation(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	allocation, err := h.portfolio.GetAllocation(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "PORTFOLIO_ALLOCATION_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, allocation)
}
