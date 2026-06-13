import { Injectable, computed, effect, inject, signal } from '@angular/core';
import { AuthService } from './auth.service';
import { PortfolioApiService } from './portfolio-api.service';
import { StockStreamService } from './stock-stream.service';
import { MAX_WATCHLIST_SIZE, WatchlistService } from './watchlist.service';

function sameSymbolList(left: string[], right: string[]): boolean {
  if (left.length !== right.length) {
    return false;
  }

  return left.every((symbol, index) => symbol === right[index]);
}

@Injectable({ providedIn: 'root' })
export class MarketStreamCoordinatorService {
  private readonly authService = inject(AuthService);
  private readonly stockStreamService = inject(StockStreamService);
  private readonly watchlistService = inject(WatchlistService);
  private readonly portfolioApi = inject(PortfolioApiService);

  private readonly active = signal(false);
  private lastConnectedSymbols: string[] = [];

  private readonly mergedSymbols = computed(
    () => {
      const monitorSymbols = this.watchlistService.symbols();
      const portfolioSymbols = this.portfolioApi.holdings().map((holding) => holding.symbol);
      const seen = new Set<string>();

      for (const symbol of [...monitorSymbols, ...portfolioSymbols]) {
        const normalized = symbol.trim().toUpperCase();
        if (normalized) {
          seen.add(normalized);
        }
      }

      return [...seen].sort();
    },
    { equal: sameSymbolList },
  );

  constructor() {
    effect(() => {
      if (!this.active() || !this.authService.currentUser()) {
        return;
      }

      const symbols = this.mergedSymbols();
      if (sameSymbolList(symbols, this.lastConnectedSymbols)) {
        return;
      }

      this.lastConnectedSymbols = symbols;

      if (symbols.length === 0) {
        this.stockStreamService.connect([]);
        return;
      }

      const limited = symbols.slice(0, MAX_WATCHLIST_SIZE);
      this.stockStreamService.connect(limited);

      if (symbols.length > MAX_WATCHLIST_SIZE) {
        this.stockStreamService.streamHint.set(
          `Streaming ${MAX_WATCHLIST_SIZE} of ${symbols.length} symbols (Finnhub free tier limit).`,
        );
      }
    });

    effect(() => {
      if (!this.authService.currentUser()) {
        this.stop();
      }
    });
  }

  async start(): Promise<void> {
    if (this.active()) {
      return;
    }

    this.active.set(true);

    try {
      await this.portfolioApi.loadHoldings();
    } catch {
      // Portfolio stream can still run with monitor symbols only.
    }
  }

  stop(): void {
    this.active.set(false);
    this.lastConnectedSymbols = [];
    this.stockStreamService.disconnect();
  }
}
