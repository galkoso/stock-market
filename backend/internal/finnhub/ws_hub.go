package finnhub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	finnhubWSURL          = "wss://ws.finnhub.io"
	maxReconnectDelay     = 30 * time.Second
	initialReconnectDelay = time.Second
	listenerChannelSize   = 64
)

type Subscription struct {
	Updates <-chan TradeUpdate
	Cancel  func()
}

type WSHub struct {
	apiKey string

	mu             sync.Mutex
	conn           *websocket.Conn
	writeMu        sync.Mutex
	symbolRefs     map[string]int
	listeners      map[string]map[uint64]chan TradeUpdate
	nextListenerID uint64
	stopped        bool
	stop           chan struct{}
	wake           chan struct{}
	wg             sync.WaitGroup
}

func NewWSHub(apiKey string) *WSHub {
	hub := &WSHub{
		apiKey:     apiKey,
		symbolRefs: make(map[string]int),
		listeners:  make(map[string]map[uint64]chan TradeUpdate),
		stop:       make(chan struct{}),
		wake:       make(chan struct{}, 1),
	}

	hub.wg.Add(1)
	go hub.connectionManager()

	return hub
}

func (h *WSHub) Subscribe(symbol string) (*Subscription, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	updates := make(chan TradeUpdate, listenerChannelSize)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return nil, fmt.Errorf("stream hub is stopped")
	}

	h.nextListenerID++
	listenerID := h.nextListenerID

	if h.listeners[normalized] == nil {
		h.listeners[normalized] = make(map[uint64]chan TradeUpdate)
	}
	h.listeners[normalized][listenerID] = updates

	isNewSymbol := h.symbolRefs[normalized] == 0
	h.symbolRefs[normalized]++

	if isNewSymbol && h.conn != nil {
		if err := h.sendLocked(finnhubSubscribeMessage{Type: "subscribe", Symbol: normalized}); err != nil {
			h.rollbackSubscribeLocked(normalized, listenerID, updates)
			return nil, fmt.Errorf("subscribe to finnhub: %w", err)
		}
	}

	h.signalWake()

	cancel := func() {
		h.unsubscribe(normalized, listenerID, updates)
	}

	return &Subscription{Updates: updates, Cancel: cancel}, nil
}

func (h *WSHub) rollbackSubscribeLocked(symbol string, listenerID uint64, updates chan TradeUpdate) {
	delete(h.listeners[symbol], listenerID)
	close(updates)
	if len(h.listeners[symbol]) == 0 {
		delete(h.listeners, symbol)
	}
	if h.symbolRefs[symbol] > 0 {
		h.symbolRefs[symbol]--
	}
	if h.symbolRefs[symbol] == 0 {
		delete(h.symbolRefs, symbol)
	}
}

func (h *WSHub) unsubscribe(symbol string, listenerID uint64, updates chan TradeUpdate) {
	h.mu.Lock()
	defer h.mu.Unlock()

	symbolListeners, ok := h.listeners[symbol]
	if !ok {
		return
	}

	delete(symbolListeners, listenerID)
	close(updates)

	if len(symbolListeners) == 0 {
		delete(h.listeners, symbol)
	}

	if h.symbolRefs[symbol] > 0 {
		h.symbolRefs[symbol]--
	}

	if h.symbolRefs[symbol] == 0 {
		delete(h.symbolRefs, symbol)
		if h.conn != nil {
			if err := h.sendLocked(finnhubSubscribeMessage{Type: "unsubscribe", Symbol: symbol}); err != nil {
				log.Printf("finnhub unsubscribe failed for %s: %v", symbol, err)
			}
		}
	}

	if len(h.symbolRefs) == 0 && h.conn != nil {
		_ = h.conn.Close()
		h.conn = nil
	}
}

func (h *WSHub) Close() error {
	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		return nil
	}
	h.stopped = true
	close(h.stop)
	conn := h.conn
	h.conn = nil
	h.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}

	h.wg.Wait()
	return nil
}

func (h *WSHub) connectionManager() {
	defer h.wg.Done()

	delay := initialReconnectDelay

	for {
		if h.isStopped() {
			return
		}

		if !h.hasActiveSubscriptions() {
			select {
			case <-h.stop:
				return
			case <-h.wake:
			}
			continue
		}

		if err := h.connectAndRead(); err != nil {
			if h.isStopped() {
				return
			}
			log.Printf("finnhub websocket disconnected: %v; reconnecting in %s", err, delay)
			time.Sleep(delay)
			delay = minDuration(delay*2, maxReconnectDelay)
			continue
		}

		delay = initialReconnectDelay
	}
}

func (h *WSHub) connectAndRead() error {
	endpoint := fmt.Sprintf("%s?token=%s", finnhubWSURL, url.QueryEscape(h.apiKey))
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return fmt.Errorf("dial finnhub websocket: %w", err)
	}

	h.mu.Lock()
	h.conn = conn
	symbols := h.activeSymbolsLocked()
	h.mu.Unlock()

	for _, symbol := range symbols {
		if err := h.send(finnhubSubscribeMessage{Type: "subscribe", Symbol: symbol}); err != nil {
			_ = conn.Close()
			return fmt.Errorf("resubscribe %s: %w", symbol, err)
		}
	}

	for {
		if h.isStopped() {
			_ = conn.Close()
			return nil
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			h.mu.Lock()
			if h.conn == conn {
				h.conn = nil
			}
			h.mu.Unlock()
			_ = conn.Close()
			return err
		}

		h.handleMessage(message)
	}
}

func (h *WSHub) handleMessage(message []byte) {
	var tradeMessage finnhubTradeMessage
	if err := json.Unmarshal(message, &tradeMessage); err != nil {
		return
	}

	if tradeMessage.Type != "trade" || len(tradeMessage.Data) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, entry := range tradeMessage.Data {
		symbol := strings.ToUpper(entry.Symbol)
		symbolListeners, ok := h.listeners[symbol]
		if !ok {
			continue
		}

		update := TradeUpdate{
			Symbol:    symbol,
			Price:     entry.Price,
			Volume:    entry.Volume,
			Timestamp: entry.Timestamp,
		}

		for _, listener := range symbolListeners {
			select {
			case listener <- update:
			default:
			}
		}
	}
}

func (h *WSHub) send(message finnhubSubscribeMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sendLocked(message)
}

func (h *WSHub) sendLocked(message finnhubSubscribeMessage) error {
	if h.conn == nil {
		return fmt.Errorf("finnhub websocket is not connected")
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	return h.conn.WriteMessage(websocket.TextMessage, payload)
}

func (h *WSHub) activeSymbolsLocked() []string {
	symbols := make([]string, 0, len(h.symbolRefs))
	for symbol := range h.symbolRefs {
		symbols = append(symbols, symbol)
	}
	return symbols
}

func (h *WSHub) hasActiveSubscriptions() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.symbolRefs) > 0
}

func (h *WSHub) signalWake() {
	select {
	case h.wake <- struct{}{}:
	default:
	}
}

func (h *WSHub) isStopped() bool {
	select {
	case <-h.stop:
		return true
	default:
		return false
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
