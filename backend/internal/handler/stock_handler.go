package handler

import (
	"errors"
	"net/http"
	"strings"

	"stock-market/backend/internal/model"
	"stock-market/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockService *service.StockService
}

func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

func (h *StockHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		respondError(c, http.StatusBadRequest, "MISSING_QUERY", "query parameter 'q' is required")
		return
	}

	quote, err := h.stockService.SearchStock(c.Request.Context(), query)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingQuery):
			respondError(c, http.StatusBadRequest, "MISSING_QUERY", err.Error())
		case errors.Is(err, service.ErrSymbolNotFound):
			respondError(c, http.StatusNotFound, "SYMBOL_NOT_FOUND", err.Error())
		case errors.Is(err, service.ErrInvalidQuote):
			respondError(c, http.StatusNotFound, "SYMBOL_NOT_FOUND", err.Error())
		default:
			if strings.Contains(err.Error(), "finnhub API error") {
				respondError(c, http.StatusBadGateway, "FINNHUB_API_ERROR", "failed to fetch data from Finnhub")
				return
			}
			respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		}
		return
	}

	c.JSON(http.StatusOK, quote)
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, model.APIError{
		Code:    code,
		Message: message,
	})
}
