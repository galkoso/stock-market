import { Injectable, NgZone, inject, isDevMode, signal } from '@angular/core';
import { LivePriceState, StreamMessage, StreamStatus } from '../models/stock.model';
import { AuthService } from './auth.service';

const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_BASE_DELAY_MS = 1000;

function symbolsKey(symbols: string[]): string {
  return [...new Set(symbols.map((symbol) => symbol.trim().toUpperCase()).filter(Boolean))].sort().join(',');
}

@Injectable({ providedIn: 'root' })
export class StockStreamService {
  private readonly zone = inject(NgZone);
  private readonly authService = inject(AuthService);

  readonly status = signal<StreamStatus>('idle');
  readonly livePrices = signal<Record<string, LivePriceState>>({});
  readonly errorMessage = signal<string | null>(null);
  readonly streamHint = signal<string | null>(null);
  readonly activeSymbols = signal<string[]>([]);

  private socket: WebSocket | null = null;
  private activeKey = '';
  private connectGeneration = 0;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private intentionalClose = false;

  connect(symbols: string[]): void {
    const normalized = [...new Set(symbols.map((s) => s.trim().toUpperCase()).filter(Boolean))].sort();
    const key = symbolsKey(normalized);

    if (normalized.length === 0) {
      this.closeSocket();
      this.activeKey = '';
      this.activeSymbols.set([]);
      this.status.set('idle');
      this.errorMessage.set(null);
      this.streamHint.set(null);
      return;
    }

    if (key === this.activeKey && this.isSocketActive()) {
      return;
    }

    if (key === this.activeKey && this.reconnectTimer) {
      return;
    }

    this.closeSocket();
    this.intentionalClose = false;
    this.activeKey = key;
    this.activeSymbols.set(normalized);
    this.status.set('connecting');
    this.errorMessage.set(null);
    this.streamHint.set(null);

    const generation = ++this.connectGeneration;
    void this.startConnection(normalized, generation);
  }

  disconnect(userInitiated = true): void {
    this.intentionalClose = userInitiated;
    this.connectGeneration += 1;
    this.clearReconnectTimer();
    this.closeSocket();
    this.reconnectAttempts = 0;
    this.activeKey = '';

    if (userInitiated) {
      this.status.set('idle');
      this.activeSymbols.set([]);
      this.livePrices.set({});
      this.errorMessage.set(null);
      this.streamHint.set(null);
    }
  }

  private async startConnection(symbols: string[], generation: number): Promise<void> {
    try {
      await this.authService.refreshAccessToken();
    } catch {
      if (generation !== this.connectGeneration) {
        return;
      }

      this.zone.run(() => {
        this.status.set('error');
        this.errorMessage.set('Unable to refresh session for live prices.');
      });
      return;
    }

    if (generation !== this.connectGeneration || symbolsKey(symbols) !== this.activeKey) {
      return;
    }

    this.openSocket(symbols, generation);
  }

  private openSocket(symbols: string[], generation: number): void {
    const url = this.buildWsUrl(symbols);
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      if (generation !== this.connectGeneration) {
        return;
      }

      this.zone.run(() => {
        this.reconnectAttempts = 0;
        this.status.set('connecting');
      });
    };

    this.socket.onmessage = (event) => {
      if (generation !== this.connectGeneration) {
        return;
      }

      this.zone.run(() => {
        this.handleMessage(event.data as string);
      });
    };

    this.socket.onerror = () => {
      if (generation !== this.connectGeneration || this.intentionalClose) {
        return;
      }

      this.zone.run(() => {
        this.status.set('error');
        this.errorMessage.set('Live price connection failed.');
      });
    };

    this.socket.onclose = () => {
      if (generation !== this.connectGeneration) {
        return;
      }

      this.zone.run(() => {
        this.socket = null;

        if (this.intentionalClose) {
          this.status.set('disconnected');
          return;
        }

        if (this.activeSymbols().length > 0) {
          this.scheduleReconnect(symbols);
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

    if (message.type === 'trade' && message.symbol && message.price != null) {
      const symbol = message.symbol.toUpperCase();
      const timestamp = message.timestamp ?? Date.now();

      this.status.set('live');
      this.livePrices.update((current) => {
        const existing = current[symbol];
        return {
          ...current,
          [symbol]: {
            price: message.price!,
            timestamp,
            tradeCount: (existing?.tradeCount ?? 0) + 1,
          },
        };
      });
      this.streamHint.set(`Live updates for ${this.activeSymbols().length} symbol(s).`);
    }
  }

  private scheduleReconnect(symbols: string[]): void {
    if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      this.status.set('error');
      this.errorMessage.set('Unable to reconnect to live prices.');
      return;
    }

    this.status.set('connecting');
    this.reconnectAttempts += 1;

    const delay = RECONNECT_BASE_DELAY_MS * this.reconnectAttempts;
    const generation = this.connectGeneration;
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.activeSymbols().length === 0 || this.intentionalClose) {
        return;
      }
      void this.startConnection(symbols, generation);
    }, delay);
  }

  private isSocketActive(): boolean {
    if (!this.socket) {
      return false;
    }

    return this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING;
  }

  private closeSocket(): void {
    if (!this.socket) {
      return;
    }

    this.socket.onopen = null;
    this.socket.onmessage = null;
    this.socket.onerror = null;
    this.socket.onclose = null;
    this.socket.close();
    this.socket = null;
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private buildWsUrl(symbols: string[]): string {
    const params = new URLSearchParams({
      symbols: symbols.join(','),
    });

    const accessToken = this.authService.getStoredAccessToken();
    if (accessToken) {
      params.set('access_token', accessToken);
    }

    if (isDevMode()) {
      return `ws://localhost:8080/ws/stocks?${params.toString()}`;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/ws/stocks?${params.toString()}`;
  }
}
