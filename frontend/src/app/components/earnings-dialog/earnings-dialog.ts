import { DatePipe } from '@angular/common';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { enrichEarningsFromHistory } from '../../lib/earnings-merge';
import { earningsReportStatus } from '../../lib/earnings-status';
import { EarningsEvent } from '../../models/market.model';
import { MarketService } from '../../services/market.service';

@Component({
  selector: 'app-earnings-dialog',
  imports: [DatePipe, RouterLink],
  templateUrl: './earnings-dialog.html',
  styleUrl: './earnings-dialog.scss',
})
export class EarningsDialog {
  private readonly marketService = inject(MarketService);

  readonly event = input.required<EarningsEvent>();
  readonly closed = output<void>();

  protected readonly detail = signal<EarningsEvent | null>(null);
  protected readonly isLoadingDetail = signal(false);
  protected readonly detailNote = signal<string | null>(null);

  protected readonly displayEvent = computed(() => this.detail() ?? this.event());
  protected readonly status = computed(() => earningsReportStatus(this.displayEvent()));
  protected readonly isUpcoming = computed(() => this.status() === 'upcoming');
  protected readonly isReported = computed(() => this.status() === 'reported');

  constructor() {
    effect(() => {
      const event = this.event();
      this.loadDetail(event);
    });
  }

  protected close(): void {
    this.closed.emit();
  }

  protected onBackdropClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      this.close();
    }
  }

  protected hourLabel(hour: string | undefined | null): string {
    switch (hour) {
      case 'bmo':
        return 'Before market open (BMO)';
      case 'amc':
        return 'After market close (AMC)';
      case 'dmh':
        return 'During market hours (DMH)';
      default:
        return 'Time TBA';
    }
  }

  protected formatRevenue(value: number | null | undefined): string {
    if (value == null) {
      return '—';
    }
    const abs = Math.abs(value);
    if (abs >= 1e9) {
      return `$${(value / 1e9).toFixed(2)}B`;
    }
    if (abs >= 1e6) {
      return `$${(value / 1e6).toFixed(1)}M`;
    }
    return `$${value.toLocaleString()}`;
  }

  protected formatEps(value: number | null | undefined): string {
    if (value == null) {
      return '—';
    }
    return value.toFixed(2);
  }

  protected formatSurprisePercent(value: number | null | undefined): string {
    if (value == null) {
      return '—';
    }
    const sign = value >= 0 ? '+' : '';
    return `${sign}${value.toFixed(2)}%`;
  }

  protected epsBeat(event: EarningsEvent): 'beat' | 'miss' | 'inline' | null {
    if (event.epsActual == null || event.epsEstimate == null) {
      return null;
    }
    const diff = event.epsActual - event.epsEstimate;
    if (Math.abs(diff) < 0.001) {
      return 'inline';
    }
    return diff > 0 ? 'beat' : 'miss';
  }

  protected revenueBeat(event: EarningsEvent): 'beat' | 'miss' | 'inline' | null {
    if (event.revenueActual == null || event.revenueEstimate == null) {
      return null;
    }
    const diff = event.revenueActual - event.revenueEstimate;
    if (Math.abs(diff) < 1) {
      return 'inline';
    }
    return diff > 0 ? 'beat' : 'miss';
  }

  private loadDetail(event: EarningsEvent): void {
    this.isLoadingDetail.set(true);
    this.detailNote.set(null);
    this.detail.set(event);

    this.marketService.getEarningsSurprises(event.symbol, 8).subscribe({
      next: (response) => {
        const history = (response.surprises ?? []).map((item) => ({
          symbol: item.symbol,
          companyName: item.symbol,
          date: item.period.slice(0, 10),
          hour: '',
          epsActual: item.epsActual,
          epsEstimate: item.epsEstimate,
          epsSurprise: item.epsSurprise,
          epsSurprisePercent: item.epsSurprisePercent,
          revenueActual: null,
          revenueEstimate: null,
          quarter: item.quarter,
          year: item.year,
        }));

        const enriched = enrichEarningsFromHistory(event, history);
        this.detail.set(enriched);

        const status = earningsReportStatus(enriched);
        if (enriched.epsActual == null && enriched.epsEstimate == null && status === 'upcoming') {
          this.detailNote.set('Finnhub calendar may not include estimates for every symbol on the free plan.');
        } else if (enriched.epsActual != null && enriched.revenueActual == null) {
          this.detailNote.set(
            'EPS actual/estimate loaded from Finnhub /stock/earnings. Revenue detail requires calendar/earnings or a paid Finnhub plan.',
          );
        }

        this.isLoadingDetail.set(false);
      },
      error: () => {
        this.detail.set(event);
        this.detailNote.set('Could not load detailed EPS history from Finnhub.');
        this.isLoadingDetail.set(false);
      },
    });
  }
}
