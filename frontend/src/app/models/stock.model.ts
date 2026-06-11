export interface StockQuote {
  symbol: string;
  companyName: string;
  currentPrice: number;
  dailyChange: number;
  dailyChangePercent: number;
  lastUpdated: string;
}

export interface QuotesResponse {
  quotes: StockQuote[];
  errors?: SymbolLookupError[];
}

export interface SymbolLookupError {
  symbol: string;
  message: string;
}

export interface ApiError {
  code: string;
  message: string;
}

export type StreamStatus = 'idle' | 'connecting' | 'live' | 'disconnected' | 'error';

export interface StreamMessage {
  type: 'status' | 'trade' | 'error';
  status?: StreamStatus;
  symbol?: string;
  symbols?: string[];
  price?: number;
  volume?: number;
  timestamp?: number;
  message?: string;
}

export interface LivePriceState {
  price: number;
  timestamp: number;
  tradeCount: number;
}

export interface WatchlistItem extends StockQuote {
  livePrice: number | null;
  liveUpdatedAt: number | null;
  tradeCount: number;
}
