package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"stock-market/backend/internal/finnhub"
	"stock-market/backend/internal/model"
	"stock-market/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	clientWriteWait      = 10 * time.Second
	clientPongWait       = 60 * time.Second
	clientPingPeriod     = 54 * time.Second
	streamMergeBufferSize = 64
)

type StockStreamHandler struct {
	stockService *service.StockService
	streamHub    *finnhub.WSHub
	upgrader     websocket.Upgrader
}

func NewStockStreamHandler(stockService *service.StockService, streamHub *finnhub.WSHub) *StockStreamHandler {
	return &StockStreamHandler{
		stockService: stockService,
		streamHub:    streamHub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "" ||
					strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "https://localhost:")
			},
		},
	}
}

func (h *StockStreamHandler) Stream(c *gin.Context) {
	symbols, err := parseSymbolsQuery(c.Query("symbol"), c.Query("symbols"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingSymbols):
			c.JSON(http.StatusBadRequest, model.APIError{
				Code:    "MISSING_SYMBOLS",
				Message: "query parameter 'symbol' or 'symbols' is required",
			})
		case errors.Is(err, service.ErrTooManySymbols):
			c.JSON(http.StatusBadRequest, model.APIError{
				Code:    "TOO_MANY_SYMBOLS",
				Message: err.Error(),
			})
		default:
			c.JSON(http.StatusBadRequest, model.APIError{
				Code:    "INVALID_SYMBOLS",
				Message: err.Error(),
			})
		}
		return
	}

	for _, symbol := range symbols {
		if err := h.stockService.ValidateSymbol(c.Request.Context(), symbol); err != nil {
			switch {
			case errors.Is(err, service.ErrSymbolNotFound):
				c.JSON(http.StatusNotFound, model.APIError{
					Code:    "SYMBOL_NOT_FOUND",
					Message: "symbol not found: " + symbol,
				})
			default:
				c.JSON(http.StatusBadGateway, model.APIError{
					Code:    "FINNHUB_API_ERROR",
					Message: "failed to validate symbol: " + symbol,
				})
			}
			return
		}
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	subscriptions, updates, cleanup, err := h.subscribeAll(symbols)
	if err != nil {
		_ = writeStreamMessage(conn, model.StreamMessage{
			Type:    "error",
			Status:  "error",
			Message: "failed to subscribe to live prices",
		})
		return
	}
	defer cleanup()

	conn.SetReadDeadline(time.Now().Add(clientPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(clientPongWait))
		return nil
	})

	_ = writeStreamMessage(conn, model.StreamMessage{
		Type:    "status",
		Status:  "connecting",
		Symbols: symbols,
	})

	done := make(chan struct{})
	go h.readPump(conn, done)

	ticker := time.NewTicker(clientPingPeriod)
	defer ticker.Stop()

	_ = writeStreamMessage(conn, model.StreamMessage{
		Type:    "status",
		Status:  "live",
		Symbols: symbols,
		Message: "Stream active for " + strings.Join(symbols, ", ") + ". Price updates on each trade.",
	})

	_ = subscriptions // kept alive until cleanup runs

	for {
		select {
		case <-done:
			_ = writeStreamMessage(conn, model.StreamMessage{
				Type:    "status",
				Status:  "disconnected",
				Symbols: symbols,
			})
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(clientWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case update, ok := <-updates:
			if !ok {
				_ = writeStreamMessage(conn, model.StreamMessage{
					Type:    "error",
					Status:  "error",
					Message: "live stream closed",
				})
				return
			}

			if err := writeStreamMessage(conn, model.StreamMessage{
				Type:      "trade",
				Status:    "live",
				Symbol:    update.Symbol,
				Price:     update.Price,
				Volume:    update.Volume,
				Timestamp: update.Timestamp,
			}); err != nil {
				return
			}
		}
	}
}

func (h *StockStreamHandler) subscribeAll(symbols []string) ([]*finnhub.Subscription, <-chan finnhub.TradeUpdate, func(), error) {
	subscriptions := make([]*finnhub.Subscription, 0, len(symbols))
	updates := make(chan finnhub.TradeUpdate, streamMergeBufferSize*len(symbols))

	for _, symbol := range symbols {
		subscription, err := h.streamHub.Subscribe(symbol)
		if err != nil {
			for _, active := range subscriptions {
				active.Cancel()
			}
			close(updates)
			return nil, nil, nil, err
		}

		subscriptions = append(subscriptions, subscription)

		go func(sub *finnhub.Subscription) {
			for update := range sub.Updates {
				updates <- update
			}
		}(subscription)
	}

	cleanup := func() {
		for _, subscription := range subscriptions {
			subscription.Cancel()
		}
		close(updates)
	}

	return subscriptions, updates, cleanup, nil
}

func (h *StockStreamHandler) readPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func writeStreamMessage(conn *websocket.Conn, message model.StreamMessage) error {
	conn.SetWriteDeadline(time.Now().Add(clientWriteWait))
	return conn.WriteJSON(message)
}
