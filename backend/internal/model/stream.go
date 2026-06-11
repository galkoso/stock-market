package model

type StreamMessage struct {
	Type      string   `json:"type"`
	Status    string   `json:"status,omitempty"`
	Symbol    string   `json:"symbol,omitempty"`
	Symbols   []string `json:"symbols,omitempty"`
	Price     float64  `json:"price,omitempty"`
	Volume    float64  `json:"volume,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
	Message   string   `json:"message,omitempty"`
}
