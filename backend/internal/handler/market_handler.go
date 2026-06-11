package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/provider/marketdata"
	"stock-market/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	market    *services.MarketService
	watchlist *services.WatchlistService
}

func NewMarketHandler(market *services.MarketService, watchlist *services.WatchlistService) *MarketHandler {
	return &MarketHandler{market: market, watchlist: watchlist}
}

func (h *MarketHandler) Search(c *gin.Context) {
	query := c.Query("q")
	results, err := h.market.Search(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "SEARCH_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *MarketHandler) Details(c *gin.Context) {
	details, err := h.market.GetDetails(c.Request.Context(), c.Param("symbol"))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "DETAILS_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

func (h *MarketHandler) Earnings(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" {
		from = time.Now().UTC().Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02")
	}

	events, err := h.market.GetEarnings(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "EARNINGS_FAILED", Message: err.Error()})
		return
	}

	if symbolsParam := c.Query("symbols"); symbolsParam != "" {
		events = filterEarningsBySymbols(events, symbolsParam)
	}

	c.JSON(http.StatusOK, gin.H{"earnings": events, "from": from, "to": to})
}

func (h *MarketHandler) EarningsSurprises(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))
	surprises, err := h.market.GetEarningsSurprises(c.Request.Context(), c.Param("symbol"), limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "EARNINGS_SURPRISES_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"surprises": surprises})
}

func (h *MarketHandler) EarningsHistory(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))

	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		c.JSON(http.StatusBadRequest, model.APIError{Code: "SYMBOLS_REQUIRED", Message: "symbols query param is required"})
		return
	}

	symbols := make([]string, 0)
	for _, symbol := range strings.Split(symbolsParam, ",") {
		trimmed := strings.ToUpper(strings.TrimSpace(symbol))
		if trimmed != "" {
			symbols = append(symbols, trimmed)
		}
	}

	events, err := h.market.GetEarningsHistory(c.Request.Context(), symbols, from, to, limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "EARNINGS_HISTORY_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"earnings": events, "from": from, "to": to})
}

func (h *MarketHandler) WatchlistEarnings(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	windowDays, _ := strconv.Atoi(c.DefaultQuery("window", "7"))
	if windowDays != 3 && windowDays != 7 && windowDays != 14 {
		windowDays = 7
	}

	symbols, err := h.watchlist.Symbols(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "WATCHLIST_FAILED", Message: err.Error()})
		return
	}

	events, err := h.market.GetWatchlistEarnings(c.Request.Context(), symbols, windowDays)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "EARNINGS_FAILED", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"windowDays": windowDays, "earnings": events})
}

func (h *MarketHandler) News(c *gin.Context) {
	articles, err := h.market.GetNews(c.Request.Context(), c.Param("symbol"))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "NEWS_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"news": articles})
}

func (h *MarketHandler) Filings(c *gin.Context) {
	filings, err := h.market.GetFilings(c.Request.Context(), c.Param("symbol"))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "FILINGS_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"filings": filings})
}

func (h *MarketHandler) Recommendations(c *gin.Context) {
	recs, err := h.market.GetRecommendations(c.Request.Context(), c.Param("symbol"))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "RECOMMENDATIONS_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": recs})
}

func (h *MarketHandler) Movers(c *gin.Context) {
	authUser, ok := auth.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: "Unauthorized"})
		return
	}

	symbols, err := h.watchlist.Symbols(c.Request.Context(), authUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIError{Code: "WATCHLIST_FAILED", Message: err.Error()})
		return
	}

	quotes, err := h.market.GetMovers(c.Request.Context(), symbols)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIError{Code: "MOVERS_FAILED", Message: err.Error()})
		return
	}

	gainers := quotes
	losers := make([]any, 0)
	if len(quotes) > 0 {
		for i := len(quotes) - 1; i >= 0 && len(losers) < 5; i-- {
			if quotes[i].DailyChangePercent < 0 {
				losers = append(losers, quotes[i])
			}
		}
		if len(gainers) > 5 {
			gainers = gainers[:5]
		}
	}

	c.JSON(http.StatusOK, gin.H{"gainers": gainers, "losers": losers})
}

func filterEarningsBySymbols(events []marketdata.EarningsEvent, symbolsParam string) []marketdata.EarningsEvent {
	symbolSet := make(map[string]struct{})
	for _, symbol := range strings.Split(symbolsParam, ",") {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized != "" {
			symbolSet[normalized] = struct{}{}
		}
	}

	filtered := make([]marketdata.EarningsEvent, 0)
	for _, event := range events {
		if _, ok := symbolSet[strings.ToUpper(event.Symbol)]; ok {
			filtered = append(filtered, event)
		}
	}

	return filtered
}
