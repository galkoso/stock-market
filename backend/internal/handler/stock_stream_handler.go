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
	clientWriteWait  = 10 * time.Second
	clientPongWait   = 60 * time.Second
	clientPingPeriod = 54 * time.Second
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
	symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, model.APIError{
			Code:    "MISSING_SYMBOL",
			Message: "query parameter 'symbol' is required",
		})
		return
	}

	if err := h.stockService.ValidateSymbol(c.Request.Context(), symbol); err != nil {
		switch {
		case errors.Is(err, service.ErrMissingQuery):
			respondError(c, http.StatusBadRequest, "MISSING_SYMBOL", "query parameter 'symbol' is required")
		case errors.Is(err, service.ErrSymbolNotFound):
			respondError(c, http.StatusNotFound, "SYMBOL_NOT_FOUND", err.Error())
		default:
			respondError(c, http.StatusBadGateway, "FINNHUB_API_ERROR", "failed to validate symbol")
		}
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	subscription, err := h.streamHub.Subscribe(symbol)
	if err != nil {
		_ = writeStreamMessage(conn, model.StreamMessage{
			Type:    "error",
			Status:  "error",
			Symbol:  symbol,
			Message: "failed to subscribe to live prices",
		})
		return
	}
	defer subscription.Cancel()

	conn.SetReadDeadline(time.Now().Add(clientPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(clientPongWait))
		return nil
	})

	_ = writeStreamMessage(conn, model.StreamMessage{
		Type:   "status",
		Status: "connecting",
		Symbol: symbol,
	})

	done := make(chan struct{})
	go h.readPump(conn, done)

	ticker := time.NewTicker(clientPingPeriod)
	defer ticker.Stop()

	_ = writeStreamMessage(conn, model.StreamMessage{
		Type:    "status",
		Status:  "live",
		Symbol:  symbol,
		Message: "Stream active. Price updates on each trade.",
	})

	for {
		select {
		case <-done:
			_ = writeStreamMessage(conn, model.StreamMessage{
				Type:   "status",
				Status: "disconnected",
				Symbol: symbol,
			})
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(clientWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case update, ok := <-subscription.Updates:
			if !ok {
				_ = writeStreamMessage(conn, model.StreamMessage{
					Type:    "error",
					Status:  "error",
					Symbol:  symbol,
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
