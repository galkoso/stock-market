package model

import "time"

type StockQuote struct {
	Symbol             string    `json:"symbol"`
	CompanyName        string    `json:"companyName"`
	CurrentPrice       float64   `json:"currentPrice"`
	DailyChange        float64   `json:"dailyChange"`
	DailyChangePercent float64   `json:"dailyChangePercent"`
	LastUpdated        time.Time `json:"lastUpdated"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
