import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, throwError } from 'rxjs';
import { NotificationRecord } from '../models/market.model';
import { AuthService } from './auth.service';

@Injectable({ providedIn: 'root' })
export class NotificationsService {
  private readonly http = inject(HttpClient);
  private readonly authService = inject(AuthService);
  private readonly apiBaseUrl = '/api';

  private eventSource: EventSource | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelayMs = 1_000;

  readonly notifications = signal<NotificationRecord[]>([]);
  readonly unreadCount = signal(0);

  async load(): Promise<void> {
    const response = await firstValueFrom(
      this.http
        .get<{ notifications: NotificationRecord[]; unreadCount: number }>(`${this.apiBaseUrl}/notifications`)
        .pipe(catchError((error) => this.handleError(error))),
    );
    this.notifications.set(response.notifications ?? []);
    this.unreadCount.set(response.unreadCount ?? 0);
  }

  connect(): void {
    const token = this.authService.getStoredAccessToken();
    if (!token) {
      return;
    }

    this.disconnect(false);

    const url = `${this.apiBaseUrl}/notifications/stream?access_token=${encodeURIComponent(token)}`;
    const source = new EventSource(url);
    this.eventSource = source;

    source.addEventListener('connected', (event) => {
      this.reconnectDelayMs = 1_000;
      const data = JSON.parse((event as MessageEvent).data) as { unreadCount?: number };
      if (typeof data.unreadCount === 'number') {
        this.unreadCount.set(data.unreadCount);
      }
    });

    source.addEventListener('notification', (event) => {
      const data = JSON.parse((event as MessageEvent).data) as {
        notification: NotificationRecord;
        unreadCount: number;
      };
      this.notifications.update((list) => [data.notification, ...list]);
      this.unreadCount.set(data.unreadCount);
    });

    source.onerror = () => {
      source.close();
      if (this.eventSource === source) {
        this.eventSource = null;
      }
      this.scheduleReconnect();
    };
  }

  disconnect(clearReconnect = true): void {
    if (clearReconnect && this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  async markRead(id: string): Promise<void> {
    await firstValueFrom(
      this.http
        .post(`${this.apiBaseUrl}/notifications/${id}/read`, {})
        .pipe(catchError((error) => this.handleError(error))),
    );

    this.notifications.update((list) =>
      list.map((item) => (item.id === id ? { ...item, isRead: true } : item)),
    );
    this.unreadCount.update((count) => Math.max(0, count - 1));
  }

  async markAllRead(): Promise<void> {
    await firstValueFrom(
      this.http
        .post(`${this.apiBaseUrl}/notifications/read-all`, {})
        .pipe(catchError((error) => this.handleError(error))),
    );

    this.notifications.update((list) => list.map((item) => ({ ...item, isRead: true })));
    this.unreadCount.set(0);
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer || !this.authService.getStoredAccessToken()) {
      return;
    }

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectDelayMs);

    this.reconnectDelayMs = Math.min(this.reconnectDelayMs * 2, 30_000);
  }

  private handleError(error: HttpErrorResponse) {
    const message =
      (error.error as { message?: string } | undefined)?.message ??
      (error.status === 0 ? 'Unable to reach the backend.' : 'Notifications request failed.');
    return throwError(() => new Error(message));
  }
}
