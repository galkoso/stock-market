import { Component, inject, OnInit, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EarningsEvent } from '../../models/market.model';
import { MarketService } from '../../services/market.service';
import { WatchlistApiService } from '../../services/watchlist-api.service';
import { WatchlistService } from '../../services/watchlist.service';

@Component({
  selector: 'app-watchlist-page',
  imports: [RouterLink],
  templateUrl: './watchlist-page.html',
  styleUrl: './watchlist-page.scss',
})
export class WatchlistPage implements OnInit {
  private readonly watchlistApi = inject(WatchlistApiService);
  private readonly marketService = inject(MarketService);
  private readonly localWatchlist = inject(WatchlistService);

  protected readonly items = this.watchlistApi.items;
  protected readonly localItems = this.localWatchlist.items;
  protected readonly earnings = signal<EarningsEvent[]>([]);
  protected readonly errorMessage = signal<string | null>(null);

  ngOnInit(): void {
    this.refresh();
  }

  protected async refresh(): Promise<void> {
    this.errorMessage.set(null);
    try {
      await this.watchlistApi.load();

      const symbols = this.localWatchlist.symbols();
      if (symbols.length === 0) {
        this.earnings.set([]);
        return;
      }

      const from = new Date().toISOString().slice(0, 10);
      const toDate = new Date();
      toDate.setDate(toDate.getDate() + 14);
      const to = toDate.toISOString().slice(0, 10);

      this.marketService.getEarnings(from, to, symbols).subscribe({
        next: (response) => this.earnings.set(response.earnings ?? []),
      });
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Failed to load watchlist');
    }
  }

  protected async remove(symbol: string): Promise<void> {
    await this.watchlistApi.remove(symbol);
    await this.refresh();
  }

  protected earningsCountdown(symbol: string): string {
    const event = this.earnings().find((e) => e.symbol === symbol);
    if (!event) return 'No earnings in next 14 days';
    const days = Math.ceil((new Date(event.date).getTime() - Date.now()) / 86400000);
    if (days <= 0) return 'Earnings today';
    if (days === 1) return 'Earnings tomorrow';
    return `Earnings in ${days} days`;
  }
}
