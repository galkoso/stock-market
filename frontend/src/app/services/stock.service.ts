import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { ApiError, StockQuote } from '../models/stock.model';

@Injectable({ providedIn: 'root' })
export class StockService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  search(query: string) {
    return this.http
      .get<StockQuote>(`${this.apiBaseUrl}/stocks/search`, {
        params: { q: query.trim() },
      })
      .pipe(
        catchError((error: HttpErrorResponse) => {
          const apiError = error.error as ApiError | undefined;
          const message =
            apiError?.message ??
            (error.status === 0
              ? 'Unable to reach the backend. Is the server running?'
              : 'Something went wrong while searching for the stock.');

          return throwError(() => new Error(message));
        }),
      );
  }
}
