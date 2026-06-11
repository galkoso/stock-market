import { EarningsEvent } from '../models/market.model';

export type EarningsReportStatus = 'upcoming' | 'reported';

export function earningsDateIso(event: EarningsEvent): string {
  return (event.date ?? '').slice(0, 10);
}

export function todayIso(): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${now.getFullYear()}-${month}-${day}`;
}

export function hasReportedResults(event: EarningsEvent): boolean {
  return event.epsActual != null || event.revenueActual != null;
}

export function earningsReportStatus(event: EarningsEvent): EarningsReportStatus {
  const date = earningsDateIso(event);
  if (hasReportedResults(event) || date < todayIso()) {
    return 'reported';
  }
  return 'upcoming';
}

export function isPastEarningsDate(dateIso: string): boolean {
  return dateIso < todayIso();
}

export function isUpcomingEarningsDate(dateIso: string): boolean {
  return dateIso >= todayIso();
}
