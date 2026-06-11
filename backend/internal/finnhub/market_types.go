package finnhub

type ProfileResponse struct {
	Country   string  `json:"country"`
	Currency  string  `json:"currency"`
	Exchange  string  `json:"exchange"`
	FinnhubIndustry string `json:"finnhubIndustry"`
	IPO       string  `json:"ipo"`
	Logo      string  `json:"logo"`
	MarketCap float64 `json:"marketCapitalization"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	ShareOutstanding float64 `json:"shareOutstanding"`
	Ticker    string  `json:"ticker"`
	WebURL    string  `json:"weburl"`
}

type EarningsCalendarResponse struct {
	EarningsCalendar []EarningsCalendarEntry `json:"earningsCalendar"`
}

type EarningsSurpriseEntry struct {
	Actual          float64 `json:"actual"`
	Estimate        float64 `json:"estimate"`
	Period          string  `json:"period"`
	Quarter         int     `json:"quarter"`
	Surprise        float64 `json:"surprise"`
	SurprisePercent float64 `json:"surprisePercent"`
	Symbol          string  `json:"symbol"`
	Year            int     `json:"year"`
}

type EarningsCalendarEntry struct {
	Date            string   `json:"date"`
	EPSActual       *float64 `json:"epsActual"`
	EPSEstimate     *float64 `json:"epsEstimate"`
	EPSSurprise     *float64 `json:"epsSurprise"`
	Hour            string   `json:"hour"`
	Quarter         int      `json:"quarter"`
	RevenueActual   *float64 `json:"revenueActual"`
	RevenueEstimate *float64 `json:"revenueEstimate"`
	Symbol          string   `json:"symbol"`
	Year            int      `json:"year"`
}

type CompanyNewsEntry struct {
	Category string `json:"category"`
	Datetime int64  `json:"datetime"`
	Headline string `json:"headline"`
	ID       int64  `json:"id"`
	Image    string `json:"image"`
	Related  string `json:"related"`
	Source   string `json:"source"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
}

type FilingEntry struct {
	AccessNumber string `json:"accessNumber"`
	Symbol       string `json:"symbol"`
	CIK          string `json:"cik"`
	Form         string `json:"form"`
	FiledDate    string `json:"filedDate"`
	AcceptedDate string `json:"acceptedDate"`
	ReportURL    string `json:"reportUrl"`
	FilingURL    string `json:"filingUrl"`
}

type RecommendationEntry struct {
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Period     string `json:"period"`
	Sell       int    `json:"sell"`
	StrongBuy  int    `json:"strongBuy"`
	StrongSell int    `json:"strongSell"`
	Symbol     string `json:"symbol"`
}
