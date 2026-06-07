import { Injectable, NgZone, inject, signal } from '@angular/core';
import { LivePriceUpdate, StreamMessage, StreamStatus } from '../models/stock.model';

const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_BASE_DELAY_MS = 1000;

@Injectable({ providedIn: 'root' })
export class StockStreamService {
  private readonly zone = inject(NgZone);

  readonly status = signal<StreamStatus>('idle');
  readonly livePrice = signal<number | null>(null);
  readonly lastTradeAt = signal<number | null>(null);
  readonly errorMessage = signal<string | null>(null);
  readonly streamHint = signal<string | null>(null);
  readonly tradeCount = signal(0);

  private socket: WebSocket | null = null;
  private activeSymbol: string | null = null;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private intentionalClose = false;

  connect(symbol: string): void {
    const normalized = symbol.trim().toUpperCase();
    if (!normalized) {
      return;
    }

    if (this.activeSymbol === normalized && this.socket) {
      return;
    }

    this.disconnect(false);
    this.intentionalClose = false;
    this.activeSymbol = normalized;
    this.status.set('connecting');
    this.errorMessage.set(null);
    this.streamHint.set(null);
    this.tradeCount.set(0);
    this.livePrice.set(null);
    this.lastTradeAt.set(null);
    this.openSocket(normalized);
  }

  disconnect(userInitiated = true): void {
    this.intentionalClose = userInitiated;
    this.clearReconnectTimer();

    if (this.socket) {
      this.socket.onopen = null;
      this.socket.onmessage = null;
      this.socket.onerror = null;
      this.socket.onclose = null;
      this.socket.close();
      this.socket = null;
    }

    this.activeSymbol = null;
    this.reconnectAttempts = 0;

    if (userInitiated) {
      this.status.set('idle');
      this.livePrice.set(null);
      this.lastTradeAt.set(null);
      this.errorMessage.set(null);
      this.streamHint.set(null);
      this.tradeCount.set(0);
    }
  }

  private openSocket(symbol: string): void {
    const url = this.buildWsUrl(symbol);
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      this.zone.run(() => {
        this.reconnectAttempts = 0;
        this.status.set('connecting');
      });
    };

    this.socket.onmessage = (event) => {
      this.zone.run(() => {
        this.handleMessage(event.data as string);
      });
    };

    this.socket.onerror = () => {
      this.zone.run(() => {
        if (!this.intentionalClose) {
          this.status.set('error');
          this.errorMessage.set('Live price connection failed.');
        }
      });
    };

    this.socket.onclose = () => {
      this.zone.run(() => {
        this.socket = null;

        if (this.intentionalClose) {
          this.status.set('disconnected');
          return;
        }

        if (this.activeSymbol === symbol) {
          this.scheduleReconnect(symbol);
        }
      });
    };
  }

  private handleMessage(raw: string): void {
    let message: StreamMessage;

    try {
      message = JSON.parse(raw) as StreamMessage;
    } catch {
      return;
    }

    if (message.type === 'status' && message.status) {
      this.status.set(message.status);
      if (message.status === 'live') {
        this.streamHint.set(
          message.message ??
            'Connected. Waiting for trades — prices tick when the US market is open.',
        );
      }
      if (message.status === 'error') {
        this.errorMessage.set(message.message ?? 'Live stream error.');
      }
      return;
    }

    if (message.type === 'error') {
      this.status.set('error');
      this.errorMessage.set(message.message ?? 'Live stream error.');
      return;
    }

    if (message.type === 'trade' && message.price != null) {
      this.status.set('live');
      this.livePrice.set(message.price);
      this.lastTradeAt.set(message.timestamp ?? Date.now());
      this.tradeCount.update((count) => count + 1);
      this.streamHint.set('Live price updating on each trade.');
    }
  }

  private scheduleReconnect(symbol: string): void {
    if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      this.status.set('error');
      this.errorMessage.set('Unable to reconnect to live prices.');
      return;
    }

    this.status.set('connecting');
    this.reconnectAttempts += 1;

    const delay = RECONNECT_BASE_DELAY_MS * this.reconnectAttempts;
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      if (this.activeSymbol === symbol && !this.intentionalClose) {
        this.openSocket(symbol);
      }
    }, delay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private buildWsUrl(symbol: string): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/ws/stocks?symbol=${encodeURIComponent(symbol)}`;
  }
}
