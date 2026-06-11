import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, throwError } from 'rxjs';
import { WatchlistItemRecord } from '../models/market.model';

@Injectable({ providedIn: 'root' })
export class WatchlistApiService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  readonly items = signal<WatchlistItemRecord[]>([]);

  async load(): Promise<void> {
    const response = await firstValueFrom(
      this.http
        .get<{ items: WatchlistItemRecord[] }>(`${this.apiBaseUrl}/watchlist`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.items.set(response.items ?? []);
  }

  async add(symbol: string, companyName: string): Promise<WatchlistItemRecord> {
    const response = await firstValueFrom(
      this.http
        .post<{ item: WatchlistItemRecord }>(`${this.apiBaseUrl}/watchlist`, { symbol, companyName })
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
    return response.item;
  }

  async remove(symbol: string): Promise<void> {
    await firstValueFrom(
      this.http
        .delete(`${this.apiBaseUrl}/watchlist/${symbol}`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
  }

  private handleError(error: HttpErrorResponse) {
    const message =
      (error.error as { message?: string } | undefined)?.message ??
      (error.status === 0 ? 'Unable to reach the backend.' : 'Watchlist request failed.');
    return throwError(() => new Error(message));
  }
}
