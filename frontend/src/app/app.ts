import { DecimalPipe, DatePipe } from '@angular/common';
import { Component, computed, inject, OnDestroy, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { StockQuote } from './models/stock.model';
import { StockStreamService } from './services/stock-stream.service';
import { StockService } from './services/stock.service';

@Component({
  selector: 'app-root',
  imports: [FormsModule, DecimalPipe, DatePipe],
  templateUrl: './app.html',
  styleUrl: './app.scss',
})
export class App implements OnDestroy {
  private readonly stockService = inject(StockService);
  private readonly stockStreamService = inject(StockStreamService);

  protected readonly searchQuery = signal('');
  protected readonly searchState = signal<'idle' | 'loading' | 'success' | 'error'>('idle');
  protected readonly stock = signal<StockQuote | null>(null);
  protected readonly errorMessage = signal<string | null>(null);

  protected readonly streamStatus = this.stockStreamService.status;
  protected readonly livePrice = this.stockStreamService.livePrice;
  protected readonly streamError = this.stockStreamService.errorMessage;
  protected readonly streamHint = this.stockStreamService.streamHint;
  protected readonly tradeCount = this.stockStreamService.tradeCount;
  protected readonly lastTradeAt = this.stockStreamService.lastTradeAt;

  protected readonly isLoading = computed(() => this.searchState() === 'loading');
  protected readonly hasResult = computed(() => this.searchState() === 'success' && this.stock() !== null);
  protected readonly hasError = computed(() => this.searchState() === 'error');
  protected readonly isPositiveChange = computed(() => (this.stock()?.dailyChange ?? 0) >= 0);

  protected readonly displayPrice = computed(() => {
    return this.livePrice() ?? this.stock()?.currentPrice ?? null;
  });

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

  ngOnDestroy(): void {
    this.stockStreamService.disconnect();
  }

  protected onSearch(event?: Event): void {
    event?.preventDefault();

    const query = this.searchQuery().trim();
    if (!query) {
      this.stockStreamService.disconnect();
      this.searchState.set('error');
      this.errorMessage.set('Enter a company name or ticker symbol.');
      this.stock.set(null);
      return;
    }

    this.stockStreamService.disconnect();
    this.searchState.set('loading');
    this.errorMessage.set(null);
    this.stock.set(null);

    this.stockService.search(query).subscribe({
      next: (quote) => {
        this.stock.set(quote);
        this.searchState.set('success');
        this.stockStreamService.connect(quote.symbol);
      },
      error: (error: Error) => {
        this.errorMessage.set(error.message);
        this.searchState.set('error');
      },
    });
  }
}
