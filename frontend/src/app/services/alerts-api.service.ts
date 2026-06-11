import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, throwError } from 'rxjs';
import { AlertRecord } from '../models/market.model';

@Injectable({ providedIn: 'root' })
export class AlertsApiService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  readonly alerts = signal<AlertRecord[]>([]);

  async load(): Promise<void> {
    const response = await firstValueFrom(
      this.http
        .get<{ alerts: AlertRecord[] }>(`${this.apiBaseUrl}/alerts`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.alerts.set(response.alerts ?? []);
  }

  async create(symbol: string, alertType: string, params: Record<string, unknown>): Promise<void> {
    await firstValueFrom(
      this.http
        .post(`${this.apiBaseUrl}/alerts`, { symbol, alertType, params })
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
    await this.evaluate();
  }

  async evaluate(): Promise<void> {
    try {
      await firstValueFrom(
        this.http.post(`${this.apiBaseUrl}/alerts/evaluate`, {}, { withCredentials: true }),
      );
    } catch {
      // Backend may still be starting; evaluation will run on the scheduler.
    }
  }

  async remove(id: string): Promise<void> {
    await firstValueFrom(
      this.http
        .delete(`${this.apiBaseUrl}/alerts/${id}`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
  }

  private handleError(error: HttpErrorResponse) {
    const message =
      (error.error as { message?: string } | undefined)?.message ??
      (error.status === 0 ? 'Unable to reach the backend.' : 'Alerts request failed.');
    return throwError(() => new Error(message));
  }
}
