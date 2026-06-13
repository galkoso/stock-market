package finnhub

type SearchResponse struct {
	Count  int            `json:"count"`
	Result []SearchResult `json:"result"`
}

type SearchResult struct {
	Description   string `json:"description"`
	DisplaySymbol string `json:"displaySymbol"`
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
}

type QuoteResponse struct {
	CurrentPrice       float64 `json:"c"`
	DailyChange        float64 `json:"d"`
	DailyChangePercent float64 `json:"dp"`
	High               float64 `json:"h"`
	Low                float64 `json:"l"`
	Open               float64 `json:"o"`
	PreviousClose      float64 `json:"pc"`
	Timestamp          int64   `json:"t"`
}

type TradeUpdate struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

type finnhubSubscribeMessage struct {
	Type   string `json:"type"`
	Symbol string `json:"symbol"`
}

type finnhubTradeMessage struct {
	Type string              `json:"type"`
	Data []finnhubTradeEntry `json:"data"`
}

type finnhubTradeEntry struct {
	Symbol    string  `json:"s"`
	Price     float64 `json:"p"`
	Volume    float64 `json:"v"`
	Timestamp int64   `json:"t"`
}
