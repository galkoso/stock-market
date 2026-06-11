import { Injectable, computed, inject, signal } from '@angular/core';
import { StockQuote, WatchlistItem } from '../models/stock.model';
import { StockStreamService } from './stock-stream.service';

const STORAGE_KEY = 'stock-market-watchlist';
const SELECTED_KEY = 'stock-market-selected';
export const MAX_WATCHLIST_SIZE = 50;

@Injectable({ providedIn: 'root' })
export class WatchlistService {
  private readonly stockStreamService = inject(StockStreamService);

  readonly items = signal<WatchlistItem[]>(this.loadFromStorage());
  readonly symbols = computed(() => this.items().map((item) => item.symbol));
  readonly selectedSymbol = signal<string | null>(this.loadSelectedFromStorage());

  readonly selectedItem = computed(() => {
    const symbol = this.selectedSymbol();
    if (!symbol) {
      return null;
    }
    return this.items().find((item) => item.symbol === symbol) ?? null;
  });

  add(quote: StockQuote): { added: boolean; message?: string } {
    const symbol = quote.symbol.toUpperCase();

    if (this.items().some((item) => item.symbol === symbol)) {
      return { added: false, message: `${symbol} is already in your watchlist.` };
    }

    if (this.items().length >= MAX_WATCHLIST_SIZE) {
      return {
        added: false,
        message: `Watchlist is full (${MAX_WATCHLIST_SIZE} symbols max on Finnhub free tier).`,
      };
    }

    const nextItem: WatchlistItem = {
      ...quote,
      livePrice: null,
      liveUpdatedAt: null,
      tradeCount: 0,
    };

    const next = [...this.items(), nextItem];
    this.items.set(next);
    this.persist(next);
    this.selectSymbol(symbol);
    this.stockStreamService.connect(next.map((item) => item.symbol));

    return { added: true };
  }

  selectSymbol(symbol: string): void {
    const normalized = symbol.toUpperCase();
    if (!this.items().some((item) => item.symbol === normalized)) {
      return;
    }
    this.selectedSymbol.set(normalized);
    localStorage.setItem(SELECTED_KEY, normalized);
  }

  clearSelection(): void {
    this.selectedSymbol.set(null);
    localStorage.removeItem(SELECTED_KEY);
  }

  remove(symbol: string): void {
    const normalized = symbol.toUpperCase();
    const next = this.items().filter((item) => item.symbol !== normalized);
    this.items.set(next);
    this.persist(next);

    if (this.selectedSymbol() === normalized) {
      if (next.length > 0) {
        this.selectSymbol(next[0].symbol);
      } else {
        this.clearSelection();
      }
    }

    if (next.length === 0) {
      this.stockStreamService.disconnect();
      return;
    }

    this.stockStreamService.connect(next.map((item) => item.symbol));
  }

  syncLivePrices(livePrices: Record<string, { price: number; timestamp: number; tradeCount: number }>): void {
    if (Object.keys(livePrices).length === 0) {
      return;
    }

    this.items.update((items) =>
      items.map((item) => {
        const live = livePrices[item.symbol];
        if (!live) {
          return item;
        }

        return {
          ...item,
          livePrice: live.price,
          liveUpdatedAt: live.timestamp,
          tradeCount: live.tradeCount,
        };
      }),
    );
  }

  restore(symbols: string[]): void {
    if (symbols.length === 0) {
      return;
    }
    this.stockStreamService.connect(symbols);
  }

  setItems(quotes: StockQuote[]): void {
    const next = quotes.map((quote) => ({
      ...quote,
      livePrice: null,
      liveUpdatedAt: null,
      tradeCount: 0,
    }));
    this.items.set(next);
    this.persist(next);
    if (next.length > 0 && !next.some((item) => item.symbol === this.selectedSymbol())) {
      this.selectSymbol(next[0].symbol);
    }
    this.stockStreamService.connect(next.map((item) => item.symbol));
  }

  private loadSelectedFromStorage(): string | null {
    const stored = localStorage.getItem(SELECTED_KEY);
    return stored ? stored.toUpperCase() : null;
  }

  private loadFromStorage(): WatchlistItem[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) {
        return [];
      }

      const parsed = JSON.parse(raw) as WatchlistItem[];
      return Array.isArray(parsed) ? parsed.slice(0, MAX_WATCHLIST_SIZE) : [];
    } catch {
      return [];
    }
  }

  private persist(items: WatchlistItem[]): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  }
}
