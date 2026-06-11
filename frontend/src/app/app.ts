import { DecimalPipe, DatePipe } from '@angular/common';
import { Component, computed, effect, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TradingViewChart } from './components/trading-view-chart/trading-view-chart';
import { WatchlistApiService } from './services/watchlist-api.service';
import { StockStreamService } from './services/stock-stream.service';
import { StockService } from './services/stock.service';
import { MAX_WATCHLIST_SIZE, WatchlistService } from './services/watchlist.service';

const CHART_OPEN_KEY = 'stock-market-chart-open';

@Component({
  selector: 'app-root',
  imports: [FormsModule, DecimalPipe, DatePipe, TradingViewChart],
  templateUrl: './app.html',
  styleUrl: './app.scss',
})
export class App implements OnInit, OnDestroy {
  private readonly stockService = inject(StockService);
  private readonly stockStreamService = inject(StockStreamService);
  private readonly watchlistService = inject(WatchlistService);
  private readonly watchlistApi = inject(WatchlistApiService);

  protected readonly searchQuery = signal('');
  protected readonly searchState = signal<'idle' | 'loading' | 'success' | 'error'>('idle');
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly sidebarMessage = signal<string | null>(null);

  protected readonly watchlist = this.watchlistService.items;
  protected readonly selectedStock = this.watchlistService.selectedItem;
  protected readonly selectedSymbol = this.watchlistService.selectedSymbol;
  protected readonly streamStatus = this.stockStreamService.status;
  protected readonly streamError = this.stockStreamService.errorMessage;
  protected readonly streamHint = this.stockStreamService.streamHint;
  protected readonly maxWatchlistSize = MAX_WATCHLIST_SIZE;
  protected readonly chartOpen = signal(loadChartOpen());

  protected readonly isLoading = computed(() => this.searchState() === 'loading');
  protected readonly hasWatchlist = computed(() => this.watchlist().length > 0);
  protected readonly hasError = computed(() => this.searchState() === 'error');

  protected readonly streamStatusLabel = computed(() => {
    switch (this.streamStatus()) {
      case 'connecting':
        return 'Connecting';
      case 'live':
        return 'Live';
      case 'disconnected':
        return 'Disconnected';
      case 'error':
        return 'Error';
      default:
        return 'Idle';
    }
  });

  constructor() {
    effect(() => {
      this.watchlistService.syncLivePrices(this.stockStreamService.livePrices());
    });
  }

  ngOnInit(): void {
    const symbols = this.watchlistService.symbols();

    const storedSelection = this.watchlistService.selectedSymbol();
    if (symbols.length > 0) {
      if (storedSelection && symbols.includes(storedSelection)) {
        this.watchlistService.selectSymbol(storedSelection);
      } else {
        this.watchlistService.selectSymbol(symbols[0]);
      }
    }

    if (symbols.length === 0) {
      return;
    }

    this.stockService.getQuotes(symbols).subscribe({
      next: (response) => {
        if (response.quotes.length > 0) {
          this.watchlistService.setItems(response.quotes);
          const current = this.watchlistService.selectedSymbol();
          if (current) {
            this.watchlistService.selectSymbol(current);
          }
        } else {
          this.watchlistService.restore(symbols);
        }
      },
      error: () => {
        this.watchlistService.restore(symbols);
      },
    });
  }

  ngOnDestroy(): void {
    this.stockStreamService.disconnect();
  }

  protected onSearch(event?: Event): void {
    event?.preventDefault();

    const query = this.searchQuery().trim();
    if (!query) {
      this.searchState.set('error');
      this.errorMessage.set('Enter a company name or ticker symbol.');
      this.sidebarMessage.set(null);
      return;
    }

    this.searchState.set('loading');
    this.errorMessage.set(null);
    this.sidebarMessage.set(null);

    this.stockService.search(query).subscribe({
      next: (quote) => {
        const result = this.watchlistService.add(quote);
        this.searchState.set('success');
        this.searchQuery.set('');

        if (result.added) {
          this.watchlistService.selectSymbol(quote.symbol);
          this.sidebarMessage.set(`${quote.symbol} added.`);
          void this.watchlistApi.add(quote.symbol, quote.companyName);
        } else {
          this.watchlistService.selectSymbol(quote.symbol);
          this.sidebarMessage.set(result.message ?? `${quote.symbol} is already in your watchlist.`);
        }
      },
      error: (error: Error) => {
        this.errorMessage.set(error.message);
        this.searchState.set('error');
      },
    });
  }

  protected selectStock(symbol: string): void {
    this.watchlistService.selectSymbol(symbol);
    this.sidebarMessage.set(null);
  }

  protected removeFromWatchlist(symbol: string, event: Event): void {
    event.stopPropagation();
    this.watchlistService.remove(symbol);
    this.sidebarMessage.set(`${symbol} removed.`);
  }

  protected displayPrice(item: { currentPrice: number; livePrice: number | null }): number {
    return item.livePrice ?? item.currentPrice;
  }

  protected isPositiveChange(dailyChange: number): boolean {
    return dailyChange >= 0;
  }

  protected updatedAt(item: { lastUpdated: string; liveUpdatedAt: number | null }): string | number {
    return item.liveUpdatedAt ?? item.lastUpdated;
  }

  protected isSelected(symbol: string): boolean {
    return this.selectedSymbol() === symbol;
  }

  protected toggleChart(): void {
    const next = !this.chartOpen();
    this.chartOpen.set(next);
    localStorage.setItem(CHART_OPEN_KEY, String(next));
  }

}

function loadChartOpen(): boolean {
  const stored = localStorage.getItem(CHART_OPEN_KEY);
  if (stored === null) {
    return false;
  }

  return stored === 'true';
}
