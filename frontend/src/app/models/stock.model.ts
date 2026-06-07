export interface StockQuote {
  symbol: string;
  companyName: string;
  currentPrice: number;
  dailyChange: number;
  dailyChangePercent: number;
  lastUpdated: string;
}

export interface ApiError {
  code: string;
  message: string;
}

export type SearchState = 'idle' | 'loading' | 'success' | 'error';

export type StreamStatus = 'idle' | 'connecting' | 'live' | 'disconnected' | 'error';

export interface StreamMessage {
  type: 'status' | 'trade' | 'error';
  status?: StreamStatus;
  symbol?: string;
  price?: number;
  volume?: number;
  timestamp?: number;
  message?: string;
}

export interface LivePriceUpdate {
  symbol: string;
  price: number;
  volume: number;
  timestamp: number;
}
