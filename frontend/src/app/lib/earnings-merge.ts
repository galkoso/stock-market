import { EarningsEvent } from '../models/market.model';

export function mergeEarningsEvents(calendar: EarningsEvent[], history: EarningsEvent[]): EarningsEvent[] {
  const merged = new Map<string, EarningsEvent>();

  for (const event of history) {
    merged.set(earningsKey(event), { ...event });
  }

  for (const event of calendar) {
    const key = earningsKey(event);
    const existing = merged.get(key);
    if (existing) {
      merged.set(key, {
        ...existing,
        ...event,
        epsActual: event.epsActual ?? existing.epsActual,
        epsEstimate: event.epsEstimate ?? existing.epsEstimate,
        epsSurprise: event.epsSurprise ?? existing.epsSurprise,
        epsSurprisePercent: event.epsSurprisePercent ?? existing.epsSurprisePercent,
        revenueActual: event.revenueActual ?? existing.revenueActual,
        revenueEstimate: event.revenueEstimate ?? existing.revenueEstimate,
        hour: event.hour || existing.hour,
        date: event.date || existing.date,
      });
    } else {
      merged.set(key, { ...event });
    }
  }

  return Array.from(merged.values()).sort((a, b) => a.date.localeCompare(b.date));
}

function earningsKey(event: EarningsEvent): string {
  return `${event.symbol.toUpperCase()}:${event.quarter}:${event.year}`;
}

export function enrichEarningsFromHistory(
  event: EarningsEvent,
  history: EarningsEvent[],
): EarningsEvent {
  const match = history.find(
    (item) =>
      item.symbol.toUpperCase() === event.symbol.toUpperCase() &&
      item.quarter === event.quarter &&
      item.year === event.year,
  );
  if (!match) {
    return event;
  }

  return {
    ...event,
    epsActual: event.epsActual ?? match.epsActual,
    epsEstimate: event.epsEstimate ?? match.epsEstimate,
    epsSurprise: event.epsSurprise ?? match.epsSurprise,
    epsSurprisePercent: event.epsSurprisePercent ?? match.epsSurprisePercent,
    date: event.date || match.date,
  };
}
