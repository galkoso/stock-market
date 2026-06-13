package marketdata

import (
	"context"
	"time"
)

type SearchResult struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"companyName"`
	Exchange    string `json:"exchange"`
	Industry    string `json:"industry"`
	Type        string `json:"type"`
}

type Quote struct {
	Symbol             string    `json:"symbol"`
	CurrentPrice       float64   `json:"currentPrice"`
	Open               float64   `json:"open"`
	High               float64   `json:"high"`
	Low                float64   `json:"low"`
	PreviousClose      float64   `json:"previousClose"`
	DailyChange        float64   `json:"dailyChange"`
	DailyChangePercent float64   `json:"dailyChangePercent"`
	LastUpdated        time.Time `json:"lastUpdated"`
}

type CompanyProfile struct {
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Exchange   string  `json:"exchange"`
	Industry   string  `json:"industry"`
	Country    string  `json:"country"`
	MarketCap  float64 `json:"marketCap"`
	Logo       string  `json:"logo"`
	WebURL     string  `json:"weburl"`
	IPO        string  `json:"ipo"`
}

type EarningsEvent struct {
	Symbol             string   `json:"symbol"`
	CompanyName        string   `json:"companyName"`
	Date               string   `json:"date"`
	Hour               string   `json:"hour"`
	EPSActual          *float64 `json:"epsActual"`
	EPSEstimate        *float64 `json:"epsEstimate"`
	EPSSurprise        *float64 `json:"epsSurprise,omitempty"`
	EPSSurprisePercent *float64 `json:"epsSurprisePercent,omitempty"`
	RevenueActual      *float64 `json:"revenueActual"`
	RevenueEstimate    *float64 `json:"revenueEstimate"`
	Quarter            int      `json:"quarter"`
	Year               int      `json:"year"`
}

type EarningsSurprise struct {
	Symbol             string   `json:"symbol"`
	Period             string   `json:"period"`
	Quarter            int      `json:"quarter"`
	Year               int      `json:"year"`
	EPSActual          float64  `json:"epsActual"`
	EPSEstimate        float64  `json:"epsEstimate"`
	EPSSurprise        float64  `json:"epsSurprise"`
	EPSSurprisePercent float64  `json:"epsSurprisePercent"`
}

type NewsArticle struct {
	Headline    string    `json:"headline"`
	Summary     string    `json:"summary"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`
}

type Filing struct {
	Form        string    `json:"form"`
	FiledDate   string    `json:"filedDate"`
	AcceptedDate time.Time `json:"acceptedDate"`
	ReportURL   string    `json:"reportUrl"`
}

type Recommendation struct {
	Buy      int `json:"buy"`
	Hold     int `json:"hold"`
	Sell     int `json:"sell"`
	StrongBuy int `json:"strongBuy"`
	StrongSell int `json:"strongSell"`
	Period   string `json:"period"`
}

type StockDetails struct {
	Symbol      string         `json:"symbol"`
	CompanyName string         `json:"companyName"`
	Quote       Quote          `json:"quote"`
	Profile     CompanyProfile `json:"profile"`
}

type Provider interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetProfile(ctx context.Context, symbol string) (*CompanyProfile, error)
	GetStockDetails(ctx context.Context, symbol string) (*StockDetails, error)
	GetEarningsCalendar(ctx context.Context, from, to string) ([]EarningsEvent, error)
	GetEarningsSurprises(ctx context.Context, symbol string, limit int) ([]EarningsSurprise, error)
	GetCompanyNews(ctx context.Context, symbol string, from, to string) ([]NewsArticle, error)
	GetFilings(ctx context.Context, symbol string) ([]Filing, error)
	GetRecommendations(ctx context.Context, symbol string) ([]Recommendation, error)
}
