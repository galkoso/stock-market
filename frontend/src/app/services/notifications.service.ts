import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, throwError } from 'rxjs';
import { NotificationRecord } from '../models/market.model';

@Injectable({ providedIn: 'root' })
export class NotificationsService {
  private readonly http = inject(HttpClient);
  private readonly apiBaseUrl = '/api';

  readonly notifications = signal<NotificationRecord[]>([]);
  readonly unreadCount = signal(0);

  async load(): Promise<void> {
    try {
      await firstValueFrom(this.http.post(`${this.apiBaseUrl}/alerts/evaluate`, {}));
    } catch {
      // Ignore if evaluate endpoint is unavailable.
    }

    const response = await firstValueFrom(
      this.http
        .get<{ notifications: NotificationRecord[]; unreadCount: number }>(`${this.apiBaseUrl}/notifications`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.notifications.set(response.notifications ?? []);
    this.unreadCount.set(response.unreadCount ?? 0);
  }

  async markRead(id: string): Promise<void> {
    await firstValueFrom(
      this.http
        .post(`${this.apiBaseUrl}/notifications/${id}/read`, {})
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
  }

  async markAllRead(): Promise<void> {
    await firstValueFrom(
      this.http
        .post(`${this.apiBaseUrl}/notifications/read-all`, {})
        .pipe(catchError((error) => this.handleError(error))),
    );
    await this.load();
  }

  private handleError(error: HttpErrorResponse) {
    const message =
      (error.error as { message?: string } | undefined)?.message ??
      (error.status === 0 ? 'Unable to reach the backend.' : 'Notifications request failed.');
    return throwError(() => new Error(message));
  }
}
