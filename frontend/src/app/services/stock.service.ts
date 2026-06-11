import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { ApiError, QuotesResponse, StockQuote } from '../models/stock.model';

@Injectable({ providedIn: 'root' })
export class StockService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  search(query: string) {
    return this.http
      .get<StockQuote>(`${this.apiBaseUrl}/stocks/search`, {
        params: { q: query.trim() },
      })
      .pipe(catchError((error) => this.handleError(error)));
  }

  getQuotes(symbols: string[]) {
    return this.http
      .get<QuotesResponse>(`${this.apiBaseUrl}/stocks/quotes`, {
        params: { symbols: symbols.join(',') },
      })
      .pipe(catchError((error) => this.handleError(error)));
  }

  private handleError(error: HttpErrorResponse) {
    const apiError = error.error as ApiError | undefined;
    const message =
      apiError?.message ??
      (error.status === 0
        ? 'Unable to reach the backend. Is the server running?'
        : 'Something went wrong while fetching stock data.');

    return throwError(() => new Error(message));
  }
}
