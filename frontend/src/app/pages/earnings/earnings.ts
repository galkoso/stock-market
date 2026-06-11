import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EarningsEvent } from '../../models/market.model';
import { MarketService } from '../../services/market.service';
import { WatchlistApiService } from '../../services/watchlist-api.service';
import { WatchlistService } from '../../services/watchlist.service';

@Component({
  selector: 'app-earnings',
  imports: [DatePipe, DecimalPipe, RouterLink],
  templateUrl: './earnings.html',
  styleUrl: './earnings.scss',
})
export class EarningsPage implements OnInit {
  private readonly marketService = inject(MarketService);
  private readonly localWatchlist = inject(WatchlistService);
  private readonly watchlistApi = inject(WatchlistApiService);

  protected readonly symbolFilter = signal('');
  protected readonly earnings = signal<EarningsEvent[]>([]);
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly isLoading = signal(false);
  protected readonly watchlistOnly = signal(true);

  private readonly watchlistSymbols = computed(() => {
    const symbols = new Set(this.localWatchlist.symbols().map((s) => s.toUpperCase()));
    for (const item of this.watchlistApi.items()) {
      symbols.add(item.symbol.toUpperCase());
    }
    return symbols;
  });

  protected readonly watchlistEmpty = computed(() => this.watchlistSymbols().size === 0);

  protected readonly filteredEarnings = computed(() => {
    const query = this.symbolFilter().trim().toUpperCase();
    let items = this.earnings();
    if (this.watchlistOnly()) {
      const symbols = this.watchlistSymbols();
      items = items.filter((item) => symbols.has(item.symbol.toUpperCase()));
    }
    if (!query) {
      return items;
    }
    return items.filter((item) => item.symbol.toUpperCase().includes(query));
  });

  ngOnInit(): void {
    this.load();
    void this.watchlistApi.load().catch(() => undefined);
  }

  protected toggleWatchlistOnly(): void {
    this.watchlistOnly.update((value) => !value);
  }

  private load(): void {
    const { from, to } = nextThreeMonthsRange();
    this.errorMessage.set(null);
    this.isLoading.set(true);

    this.marketService.getEarnings(from, to).subscribe({
      next: (response) => {
        this.earnings.set(response.earnings ?? []);
        this.isLoading.set(false);
      },
      error: (error: Error) => {
        this.earnings.set([]);
        this.errorMessage.set(error.message);
        this.isLoading.set(false);
      },
    });
  }
}

function nextThreeMonthsRange(): { from: string; to: string } {
  const now = new Date();
  const end = new Date(now);
  end.setMonth(end.getMonth() + 3);
  return { from: toIsoDate(now), to: toIsoDate(end) };
}

function toIsoDate(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${month}-${day}`;
}
