import { DecimalPipe, DatePipe } from '@angular/common';
import { Component, computed, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { PortfolioAllocationChart } from '../../components/portfolio-allocation-chart/portfolio-allocation-chart';
import { mergeLiveAllocation, livePortfolioTotals, sortPortfolioRows, defaultSortDirection, PortfolioSortColumn, PortfolioSortDirection } from '../../lib/portfolio-live';
import { PortfolioHoldingRecord } from '../../models/market.model';
import { PortfolioApiService } from '../../services/portfolio-api.service';
import { StockService } from '../../services/stock.service';
import { StockStreamService } from '../../services/stock-stream.service';
import { WatchlistApiService } from '../../services/watchlist-api.service';
import { WatchlistService } from '../../services/watchlist.service';

const REFRESH_INTERVAL_MS = 30_000;

@Component({
  selector: 'app-portfolio-page',
  imports: [FormsModule, DecimalPipe, DatePipe, PortfolioAllocationChart],
  templateUrl: './portfolio-page.html',
  styleUrl: './portfolio-page.scss',
})
export class PortfolioPage implements OnInit, OnDestroy {
  private readonly portfolioApi = inject(PortfolioApiService);
  private readonly stockStreamService = inject(StockStreamService);
  private readonly stockService = inject(StockService);
  private readonly watchlistService = inject(WatchlistService);
  private readonly watchlistApi = inject(WatchlistApiService);
  private refreshTimer: ReturnType<typeof setInterval> | null = null;

  protected readonly holdings = this.portfolioApi.holdings;
  protected readonly allocation = this.portfolioApi.allocation;
  protected readonly streamStatus = this.stockStreamService.status;
  protected readonly streamError = this.stockStreamService.errorMessage;
  protected readonly livePrices = this.stockStreamService.livePrices;

  protected readonly symbol = signal('');
  protected readonly quantity = signal('');
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly monitorMessage = signal<string | null>(null);
  protected readonly isSubmitting = signal(false);
  protected readonly addingToMonitor = signal<string | null>(null);
  protected readonly isLoading = signal(true);
  protected readonly editingSymbol = signal<string | null>(null);
  protected readonly editQuantity = signal('');
  protected readonly sortColumn = signal<PortfolioSortColumn | null>(null);
  protected readonly sortDirection = signal<PortfolioSortDirection>('desc');

  protected readonly usdToIls = computed(() => this.allocation().usdToIls);
  protected readonly liveAllocationHoldings = computed(() =>
    mergeLiveAllocation(this.allocation().holdings, this.livePrices(), this.usdToIls()),
  );
  protected readonly sortedTableRows = computed(() => {
    const liveBySymbol = new Map(this.liveAllocationHoldings().map((row) => [row.symbol, row]));
    const rows = this.holdings().map((holding) => ({
      holding,
      live: liveBySymbol.get(holding.symbol),
    }));

    const column = this.sortColumn();
    if (!column) {
      return rows;
    }

    return sortPortfolioRows(rows, column, this.sortDirection());
  });
  protected readonly totalValue = computed(() => livePortfolioTotals(this.liveAllocationHoldings(), this.usdToIls()).totalValue);
  protected readonly totalValueIls = computed(() => livePortfolioTotals(this.liveAllocationHoldings(), this.usdToIls()).totalValueIls);
  protected readonly dailyPnL = computed(() => livePortfolioTotals(this.liveAllocationHoldings(), this.usdToIls()).dailyPnL);
  protected readonly dailyPnLIls = computed(() => livePortfolioTotals(this.liveAllocationHoldings(), this.usdToIls()).dailyPnLIls);
  protected readonly holdingsCount = computed(() => this.holdings().length);
  protected readonly monitorSymbols = this.watchlistService.symbols;
  protected readonly hasHoldings = computed(() => this.holdings().length > 0);
  protected readonly lastUpdated = signal<Date | null>(null);

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

  ngOnInit(): void {
    void this.bootstrap();
    this.refreshTimer = setInterval(() => {
      void this.refreshAllocation();
    }, REFRESH_INTERVAL_MS);
  }

  ngOnDestroy(): void {
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
    }
  }

  protected async addHolding(): Promise<void> {
    const normalizedSymbol = this.symbol().trim().toUpperCase();
    const parsedQuantity = Number(this.quantity());

    if (!normalizedSymbol) {
      this.errorMessage.set('Enter a ticker symbol.');
      return;
    }
    if (!Number.isFinite(parsedQuantity) || parsedQuantity <= 0) {
      this.errorMessage.set('Quantity must be greater than zero.');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);

    try {
      await this.portfolioApi.add(normalizedSymbol, parsedQuantity);
      this.symbol.set('');
      this.quantity.set('');
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Unable to add holding.');
    } finally {
      this.isSubmitting.set(false);
    }
  }

  protected startEdit(holding: PortfolioHoldingRecord): void {
    this.editingSymbol.set(holding.symbol);
    this.editQuantity.set(String(holding.quantity));
    this.errorMessage.set(null);
  }

  protected cancelEdit(): void {
    this.editingSymbol.set(null);
    this.editQuantity.set('');
  }

  protected async saveEdit(symbol: string): Promise<void> {
    const parsedQuantity = Number(this.editQuantity());
    if (!Number.isFinite(parsedQuantity) || parsedQuantity <= 0) {
      this.errorMessage.set('Quantity must be greater than zero.');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);

    try {
      await this.portfolioApi.updateQuantity(symbol, parsedQuantity);
      this.cancelEdit();
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Unable to update holding.');
    } finally {
      this.isSubmitting.set(false);
    }
  }

  protected async remove(symbol: string): Promise<void> {
    this.errorMessage.set(null);
    try {
      await this.portfolioApi.remove(symbol);
      if (this.editingSymbol() === symbol) {
        this.cancelEdit();
      }
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Unable to remove holding.');
    }
  }

  protected isInMonitor(symbol: string): boolean {
    return this.monitorSymbols().includes(symbol.toUpperCase());
  }

  protected addToMonitor(symbol: string): void {
    const normalized = symbol.toUpperCase();
    if (this.isInMonitor(normalized)) {
      this.monitorMessage.set(`${normalized} is already in Monitor.`);
      return;
    }

    this.addingToMonitor.set(normalized);
    this.monitorMessage.set(null);

    this.stockService.getQuotes([normalized]).subscribe({
      next: (response) => {
        const quote = response.quotes.find((item) => item.symbol === normalized) ?? response.quotes[0];
        if (!quote) {
          this.monitorMessage.set(`Unable to load quote for ${normalized}.`);
          this.addingToMonitor.set(null);
          return;
        }

        const result = this.watchlistService.add(quote);
        if (result.added) {
          void this.watchlistApi.add(quote.symbol, quote.companyName);
          this.monitorMessage.set(`${normalized} added to Monitor.`);
        } else {
          this.monitorMessage.set(result.message ?? `${normalized} is already in Monitor.`);
        }

        this.addingToMonitor.set(null);
      },
      error: (error: Error) => {
        this.monitorMessage.set(error.message);
        this.addingToMonitor.set(null);
      },
    });
  }

  protected toggleSort(column: PortfolioSortColumn): void {
    if (this.sortColumn() === column) {
      this.sortDirection.set(this.sortDirection() === 'asc' ? 'desc' : 'asc');
      return;
    }

    this.sortColumn.set(column);
    this.sortDirection.set(defaultSortDirection(column));
  }

  protected sortIndicator(column: PortfolioSortColumn): string | null {
    if (this.sortColumn() !== column) {
      return null;
    }
    return this.sortDirection() === 'asc' ? '↑' : '↓';
  }

  protected isPositiveChange(value: number): boolean {
    return value >= 0;
  }

  private async bootstrap(): Promise<void> {
    this.isLoading.set(true);
    try {
      await this.portfolioApi.refresh();
      this.lastUpdated.set(new Date());
    } catch (error) {
      this.errorMessage.set(error instanceof Error ? error.message : 'Unable to load portfolio.');
    } finally {
      this.isLoading.set(false);
    }
  }

  private async refreshAllocation(): Promise<void> {
    try {
      await this.portfolioApi.loadAllocation();
      this.lastUpdated.set(new Date());
    } catch {
      // Keep previous allocation on transient refresh failures.
    }
  }
}
