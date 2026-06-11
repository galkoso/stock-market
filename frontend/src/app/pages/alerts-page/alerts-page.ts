import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, effect, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AlertRecord } from '../../models/market.model';
import { StockQuote } from '../../models/stock.model';
import { AlertsApiService } from '../../services/alerts-api.service';
import { StockService } from '../../services/stock.service';

const ALERT_TYPE_LABELS: Record<string, string> = {
  earnings_days: 'Earnings reminder',
  price_above: 'Price above target',
  price_below: 'Price below target',
  new_filing: 'New SEC filing',
  unusual_move: 'Unusual price move',
};

@Component({
  selector: 'app-alerts-page',
  imports: [FormsModule, DatePipe, DecimalPipe],
  templateUrl: './alerts-page.html',
  styleUrl: './alerts-page.scss',
})
export class AlertsPage implements OnInit, OnDestroy {
  private readonly alertsApi = inject(AlertsApiService);
  private readonly stockService = inject(StockService);

  private pricePollTimer: ReturnType<typeof setInterval> | null = null;

  protected readonly alerts = this.alertsApi.alerts;
  protected readonly prices = signal<Record<string, StockQuote>>({});
  protected readonly symbol = signal('');
  protected readonly alertType = signal('earnings_days');
  protected readonly targetPrice = signal('');
  protected readonly earningsDays = signal('3');
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly isSubmitting = signal(false);

  constructor() {
    effect(() => {
      const symbols = [...new Set(
        this.alerts()
          .map((alert) => alert.symbol?.trim().toUpperCase())
          .filter((symbol): symbol is string => !!symbol),
      )];
      this.loadPrices(symbols);
    });
  }

  ngOnInit(): void {
    void this.alertsApi.load().then(() => this.alertsApi.evaluate());
    this.pricePollTimer = setInterval(() => {
      const symbols = [...new Set(
        this.alerts()
          .map((alert) => alert.symbol?.trim().toUpperCase())
          .filter((symbol): symbol is string => !!symbol),
      )];
      this.loadPrices(symbols);
    }, 60_000);
  }

  ngOnDestroy(): void {
    if (this.pricePollTimer) {
      clearInterval(this.pricePollTimer);
    }
  }

  protected quoteFor(symbol: string | undefined): StockQuote | null {
    if (!symbol) {
      return null;
    }
    return this.prices()[symbol.toUpperCase()] ?? null;
  }

  protected alertTypeLabel(type: string): string {
    return ALERT_TYPE_LABELS[type] ?? type;
  }

  protected alertTypeClass(type: string): string {
    switch (type) {
      case 'earnings_days':
        return 'earnings';
      case 'price_above':
      case 'price_below':
        return 'price';
      case 'new_filing':
        return 'filing';
      case 'unusual_move':
        return 'move';
      default:
        return '';
    }
  }

  protected alertDetail(alert: AlertRecord): string {
    const params = alert.params ?? {};
    switch (alert.alertType) {
      case 'earnings_days': {
        const days = params['days'];
        return `Notify ${days ?? '?'} day(s) before earnings`;
      }
      case 'price_above': {
        const price = params['price'];
        return `When price goes above $${formatParamNumber(price)}`;
      }
      case 'price_below': {
        const price = params['price'];
        return `When price drops below $${formatParamNumber(price)}`;
      }
      case 'new_filing':
        return 'When a new SEC filing is published';
      case 'unusual_move':
        return 'When an unusual price move is detected';
      default:
        return alert.alertType;
    }
  }

  protected async createAlert(): Promise<void> {
    this.errorMessage.set(null);
    this.isSubmitting.set(true);

    const params: Record<string, unknown> = {};
    if (this.alertType() === 'price_above' || this.alertType() === 'price_below') {
      const price = Number(this.targetPrice());
      if (!Number.isFinite(price) || price <= 0) {
        this.errorMessage.set('Enter a valid target price.');
        this.isSubmitting.set(false);
        return;
      }
      params['price'] = price;
    }
    if (this.alertType() === 'earnings_days') {
      const days = Number(this.earningsDays());
      if (!Number.isFinite(days) || days < 1) {
        this.errorMessage.set('Enter at least 1 day before earnings.');
        this.isSubmitting.set(false);
        return;
      }
      params['days'] = days;
    }

    try {
      await this.alertsApi.create(this.symbol(), this.alertType(), params);
      this.symbol.set('');
      this.targetPrice.set('');
      this.earningsDays.set('3');
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Failed to create alert');
    } finally {
      this.isSubmitting.set(false);
    }
  }

  protected async remove(id: string): Promise<void> {
    await this.alertsApi.remove(id);
  }

  private loadPrices(symbols: string[]): void {
    if (symbols.length === 0) {
      this.prices.set({});
      return;
    }

    this.stockService.getQuotes(symbols).subscribe({
      next: (response) => {
        const map: Record<string, StockQuote> = {};
        for (const quote of response.quotes ?? []) {
          map[quote.symbol.toUpperCase()] = quote;
        }
        this.prices.set(map);
      },
      error: () => {
        // Keep last known prices on refresh failure.
      },
    });
  }
}

function formatParamNumber(value: unknown): string {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return '—';
  }
  return num.toFixed(2);
}
