import { DecimalPipe, DatePipe } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TradingViewChart } from '../../components/trading-view-chart/trading-view-chart';
import { EarningsEvent, Filing, NewsArticle, Recommendation, StockDetails } from '../../models/market.model';
import { MarketService } from '../../services/market.service';

@Component({
  selector: 'app-stock-details',
  imports: [DecimalPipe, DatePipe, RouterLink, TradingViewChart],
  templateUrl: './stock-details.html',
  styleUrl: './stock-details.scss',
})
export class StockDetailsPage implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly marketService = inject(MarketService);

  protected readonly details = signal<StockDetails | null>(null);
  protected readonly news = signal<NewsArticle[]>([]);
  protected readonly filings = signal<Filing[]>([]);
  protected readonly recommendations = signal<Recommendation[]>([]);
  protected readonly earnings = signal<EarningsEvent[]>([]);
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly chartOpen = signal(true);

  ngOnInit(): void {
    const symbol = this.route.snapshot.paramMap.get('symbol') ?? '';
    this.load(symbol);
  }

  private load(symbol: string): void {
    this.errorMessage.set(null);
    this.marketService.getDetails(symbol).subscribe({
      next: (details) => this.details.set(details),
      error: (error: Error) => this.errorMessage.set(error.message),
    });

    this.marketService.getNews(symbol).subscribe({
      next: (response) => this.news.set(response.news ?? []),
    });

    this.marketService.getFilings(symbol).subscribe({
      next: (response) => this.filings.set(response.filings ?? []),
    });

    this.marketService.getRecommendations(symbol).subscribe({
      next: (response) => this.recommendations.set(response.recommendations ?? []),
    });

    const from = new Date().toISOString().slice(0, 10);
    const toDate = new Date();
    toDate.setDate(toDate.getDate() + 30);
    const to = toDate.toISOString().slice(0, 10);
    this.marketService.getEarnings(from, to).subscribe({
      next: (response) => {
        const filtered = (response.earnings ?? []).filter((e) => e.symbol === symbol.toUpperCase());
        this.earnings.set(filtered);
      },
    });
  }
}
