package model

type QuotesResponse struct {
	Quotes []StockQuote      `json:"quotes"`
	Errors []SymbolLookupErr `json:"errors,omitempty"`
}

type SymbolLookupErr struct {
	Symbol  string `json:"symbol"`
	Message string `json:"message"`
}
