import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import {
  EarningsEvent,
  EarningsSurprise,
  Filing,
  NewsArticle,
  Recommendation,
  SearchResult,
  StockDetails,
  StockQuoteDetails,
} from '../models/market.model';

@Injectable({ providedIn: 'root' })
export class MarketService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  search(query: string) {
    return this.http
      .get<{ results: SearchResult[] }>(`${this.apiBaseUrl}/market/search`, { params: { q: query } })
      .pipe(catchError((error) => this.handleError(error)));
  }

  getDetails(symbol: string) {
    return this.http
      .get<StockDetails>(`${this.apiBaseUrl}/stocks/${symbol}`)
      .pipe(catchError((error) => this.handleError(error)));
  }

  getEarnings(from: string, to: string, symbols?: string[]) {
    const params: Record<string, string> = { from, to };
    if (symbols && symbols.length > 0) {
      params['symbols'] = symbols.join(',');
    }
    return this.http
      .get<{ earnings: EarningsEvent[] }>(`${this.apiBaseUrl}/earnings`, { params })
      .pipe(catchError((error) => this.handleError(error)));
  }

  getEarningsHistory(from: string, to: string, symbols: string[], limit = 8) {
    return this.http
      .get<{ earnings: EarningsEvent[] }>(`${this.apiBaseUrl}/earnings/history`, {
        params: { from, to, symbols: symbols.join(','), limit: String(limit) },
      })
      .pipe(catchError((error) => this.handleError(error)));
  }

  getEarningsSurprises(symbol: string, limit = 8) {
    return this.http
      .get<{ surprises: EarningsSurprise[] }>(
        `${this.apiBaseUrl}/stocks/${symbol}/earnings-surprises`,
        { params: { limit: String(limit) } },
      )
      .pipe(catchError((error) => this.handleError(error)));
  }

  getWatchlistEarnings(windowDays: 3 | 7 | 14) {
    return this.http
      .get<{ earnings: EarningsEvent[]; windowDays: number }>(`${this.apiBaseUrl}/watchlist/earnings`, {
        params: { window: String(windowDays) },
      })
      .pipe(catchError((error) => this.handleError(error)));
  }

  getNews(symbol: string) {
    return this.http
      .get<{ news: NewsArticle[] }>(`${this.apiBaseUrl}/news/${symbol}`)
      .pipe(catchError((error) => this.handleError(error)));
  }

  getFilings(symbol: string) {
    return this.http
      .get<{ filings: Filing[] }>(`${this.apiBaseUrl}/filings/${symbol}`)
      .pipe(catchError((error) => this.handleError(error)));
  }

  getRecommendations(symbol: string) {
    return this.http
      .get<{ recommendations: Recommendation[] }>(`${this.apiBaseUrl}/stocks/${symbol}/recommendations`)
      .pipe(catchError((error) => this.handleError(error)));
  }

  getMovers() {
    return this.http
      .get<{ gainers: StockQuoteDetails[]; losers: StockQuoteDetails[] }>(`${this.apiBaseUrl}/movers`)
      .pipe(catchError((error) => this.handleError(error)));
  }

  private handleError(error: HttpErrorResponse) {
    let message = (error.error as { message?: string } | undefined)?.message;
    if (!message) {
      if (error.status === 0) {
        message = 'Unable to reach the backend. Restart with: npm start';
      } else if (error.status === 404) {
        message = 'API not found — restart the backend (npm start).';
      } else {
        message = 'Market data request failed.';
      }
    }
    return throwError(() => new Error(message));
  }
}
