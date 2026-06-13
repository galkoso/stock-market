import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, throwError } from 'rxjs';
import {
  AllocationHoldingRecord,
  PortfolioAllocationRecord,
  PortfolioHoldingRecord,
} from '../models/market.model';

@Injectable({ providedIn: 'root' })
export class PortfolioApiService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  readonly holdings = signal<PortfolioHoldingRecord[]>([]);
  readonly allocation = signal<PortfolioAllocationRecord>({
    totalValue: 0,
    totalValueIls: 0,
    usdToIls: 0,
    holdings: [],
  });

  async loadHoldings(): Promise<void> {
    const response = await firstValueFrom(
      this.http
        .get<{ holdings: PortfolioHoldingRecord[] }>(`${this.apiBaseUrl}/portfolio`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.holdings.set(response.holdings ?? []);
  }

  async loadAllocation(): Promise<PortfolioAllocationRecord> {
    const response = await firstValueFrom(
      this.http
        .get<PortfolioAllocationRecord>(`${this.apiBaseUrl}/portfolio/allocation`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.allocation.set(response);
    return response;
  }

  async add(symbol: string, quantity: number): Promise<PortfolioHoldingRecord> {
    const response = await firstValueFrom(
      this.http
        .post<{ holding: PortfolioHoldingRecord }>(`${this.apiBaseUrl}/portfolio`, { symbol, quantity })
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.refresh();
    return response.holding;
  }

  async updateQuantity(symbol: string, quantity: number): Promise<PortfolioHoldingRecord> {
    const response = await firstValueFrom(
      this.http
        .put<{ holding: PortfolioHoldingRecord }>(`${this.apiBaseUrl}/portfolio/${symbol}`, { quantity })
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.refresh();
    return response.holding;
  }

  async remove(symbol: string): Promise<void> {
    await firstValueFrom(
      this.http
        .delete(`${this.apiBaseUrl}/portfolio/${symbol}`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.refresh();
  }

  async refresh(): Promise<void> {
    await Promise.all([this.loadHoldings(), this.loadAllocation()]);
  }

  allocationForSymbol(symbol: string): AllocationHoldingRecord | undefined {
    return this.allocation().holdings.find((holding) => holding.symbol === symbol);
  }

  private handleError(error: HttpErrorResponse) {
    const message =
      (error.error as { message?: string } | undefined)?.message ??
      (error.status === 0 ? 'Unable to reach the backend.' : 'Portfolio request failed.');
    return throwError(() => new Error(message));
  }
}
