import { AllocationHoldingRecord, LiveAllocationHoldingRecord } from '../models/market.model';
import { LivePriceState } from '../models/stock.model';

export function mergeLiveAllocation(
  holdings: AllocationHoldingRecord[],
  livePrices: Record<string, LivePriceState>,
  usdToIls: number,
): LiveAllocationHoldingRecord[] {
  const enriched = holdings.map((holding) => {
    const live = livePrices[holding.symbol];
    const price = live?.price ?? holding.price;
    const isLive = live != null;
    const previousClose = holding.previousClose;
    const dailyChange =
      previousClose > 0 ? price - previousClose : isLive ? price - holding.price + holding.dailyChange : holding.dailyChange;
    const dailyChangePercent =
      previousClose > 0 ? (dailyChange / previousClose) * 100 : holding.dailyChangePercent;
    const dailyPnL = holding.quantity * dailyChange;
    const dailyPnLIls = dailyPnL * usdToIls;

    return {
      ...holding,
      price,
      priceIls: price * usdToIls,
      dailyChange,
      dailyChangePercent,
      dailyPnL,
      dailyPnLIls,
      marketValue: holding.quantity * price,
      marketValueIls: holding.quantity * price * usdToIls,
      livePrice: live?.price ?? null,
      isLive,
    };
  });

  const totalValue = enriched.reduce((sum, holding) => sum + holding.marketValue, 0);

  const withAllocation = enriched.map((holding) => ({
    ...holding,
    allocationPercent: totalValue > 0 ? Math.round((holding.marketValue / totalValue) * 10000) / 100 : 0,
  }));

  return withAllocation.sort((a, b) => b.marketValue - a.marketValue);
}

export function livePortfolioTotals(holdings: LiveAllocationHoldingRecord[], usdToIls: number) {
  const totalValue = holdings.reduce((sum, holding) => sum + holding.marketValue, 0);
  const dailyPnL = holdings.reduce((sum, holding) => sum + holding.dailyPnL, 0);
  return {
    totalValue,
    totalValueIls: totalValue * usdToIls,
    dailyPnL,
    dailyPnLIls: dailyPnL * usdToIls,
  };
}

export type PortfolioSortColumn =
  | 'symbol'
  | 'dailyPnL'
  | 'price'
  | 'dailyChangePercent'
  | 'marketValue'
  | 'allocationPercent'
  | 'updatedAt';

export type PortfolioSortDirection = 'asc' | 'desc';

export interface PortfolioTableRow {
  holding: { id: string; symbol: string; quantity: number; updatedAt: string };
  live?: LiveAllocationHoldingRecord;
}

export function sortPortfolioRows(
  rows: PortfolioTableRow[],
  column: PortfolioSortColumn,
  direction: PortfolioSortDirection,
): PortfolioTableRow[] {
  const multiplier = direction === 'asc' ? 1 : -1;

  return [...rows].sort((a, b) => {
    const left = sortValue(column, a);
    const right = sortValue(column, b);

    if (typeof left === 'string' && typeof right === 'string') {
      return left.localeCompare(right) * multiplier;
    }

    return ((left as number) - (right as number)) * multiplier;
  });
}

function sortValue(column: PortfolioSortColumn, row: PortfolioTableRow): string | number {
  const live = row.live;

  switch (column) {
    case 'symbol':
      return row.holding.symbol;
    case 'dailyPnL':
      return live?.dailyPnL ?? Number.NEGATIVE_INFINITY;
    case 'price':
      return live?.price ?? Number.NEGATIVE_INFINITY;
    case 'dailyChangePercent':
      return live?.dailyChangePercent ?? Number.NEGATIVE_INFINITY;
    case 'marketValue':
      return live?.marketValue ?? Number.NEGATIVE_INFINITY;
    case 'allocationPercent':
      return live?.allocationPercent ?? Number.NEGATIVE_INFINITY;
    case 'updatedAt':
      return new Date(row.holding.updatedAt).getTime();
    default:
      return 0;
  }
}

export function defaultSortDirection(column: PortfolioSortColumn): PortfolioSortDirection {
  return column === 'symbol' || column === 'updatedAt' ? 'asc' : 'desc';
}
