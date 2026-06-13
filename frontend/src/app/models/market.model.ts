export interface SearchResult {
  symbol: string;
  companyName: string;
  exchange: string;
  industry: string;
  type: string;
}

export interface StockQuoteDetails {
  symbol: string;
  currentPrice: number;
  open: number;
  high: number;
  low: number;
  previousClose: number;
  dailyChange: number;
  dailyChangePercent: number;
  lastUpdated: string;
}

export interface CompanyProfile {
  symbol: string;
  name: string;
  exchange: string;
  industry: string;
  country: string;
  marketCap: number;
  logo: string;
  weburl: string;
  ipo: string;
}

export interface StockDetails {
  symbol: string;
  companyName: string;
  quote: StockQuoteDetails;
  profile: CompanyProfile;
}

export interface EarningsEvent {
  symbol: string;
  companyName: string;
  date: string;
  hour: string;
  epsActual: number | null;
  epsEstimate: number | null;
  epsSurprise?: number | null;
  epsSurprisePercent?: number | null;
  revenueActual: number | null;
  revenueEstimate: number | null;
  quarter: number;
  year: number;
}

export interface EarningsSurprise {
  symbol: string;
  period: string;
  quarter: number;
  year: number;
  epsActual: number;
  epsEstimate: number;
  epsSurprise: number;
  epsSurprisePercent: number;
}

export interface NewsArticle {
  headline: string;
  summary: string;
  source: string;
  publishedAt: string;
  url: string;
}

export interface Filing {
  form: string;
  filedDate: string;
  acceptedDate: string;
  reportUrl: string;
}

export interface Recommendation {
  buy: number;
  hold: number;
  sell: number;
  strongBuy: number;
  strongSell: number;
  period: string;
}

export interface WatchlistItemRecord {
  id: string;
  userId: string;
  symbol: string;
  companyName: string;
  createdAt: string;
}

export interface AlertRecord {
  id: string;
  userId: string;
  symbol?: string;
  alertType: string;
  params: Record<string, unknown>;
  isActive: boolean;
  lastTriggeredAt?: string;
  createdAt: string;
}

export interface NotificationRecord {
  id: string;
  userId: string;
  alertId?: string;
  symbol?: string;
  title: string;
  message: string;
  isRead: boolean;
  createdAt: string;
}

export interface PortfolioHoldingRecord {
  id: string;
  userId: string;
  symbol: string;
  quantity: number;
  createdAt: string;
  updatedAt: string;
}

export interface AllocationHoldingRecord {
  symbol: string;
  quantity: number;
  price: number;
  priceIls: number;
  previousClose: number;
  dailyChange: number;
  dailyChangePercent: number;
  marketValue: number;
  marketValueIls: number;
  allocationPercent: number;
}

export interface LiveAllocationHoldingRecord extends AllocationHoldingRecord {
  livePrice: number | null;
  isLive: boolean;
  dailyPnL: number;
  dailyPnLIls: number;
}

export interface PortfolioAllocationRecord {
  totalValue: number;
  totalValueIls: number;
  usdToIls: number;
  holdings: AllocationHoldingRecord[];
}
