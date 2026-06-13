package finnhub

// StreamClient defines live price streaming through the shared Finnhub WebSocket hub.
type StreamClient interface {
	Subscribe(symbol string) (*Subscription, error)
	Close() error
}
