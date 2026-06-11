import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, computed, HostListener, inject, OnInit, signal } from '@angular/core';
import { EarningsDialog } from '../../components/earnings-dialog/earnings-dialog';
import { earningsDateIso, isPastEarningsDate, todayIso } from '../../lib/earnings-status';
import { mergeEarningsEvents } from '../../lib/earnings-merge';
import { EarningsEvent } from '../../models/market.model';
import { MarketService } from '../../services/market.service';
import { WatchlistApiService } from '../../services/watchlist-api.service';
import { WatchlistService } from '../../services/watchlist.service';

interface CalendarDay {
  iso: string;
  dayNumber: number;
  inMonth: boolean;
  isToday: boolean;
  events: EarningsEvent[];
  visibleEvents: EarningsEvent[];
  moreCount: number;
}

const MAX_CHIPS_PER_DAY = 3;

@Component({
  selector: 'app-calendar-page',
  imports: [DatePipe, DecimalPipe, EarningsDialog],
  templateUrl: './calendar-page.html',
  styleUrl: './calendar-page.scss',
})
export class CalendarPage implements OnInit {
  private readonly marketService = inject(MarketService);
  private readonly localWatchlist = inject(WatchlistService);
  private readonly watchlistApi = inject(WatchlistApiService);

  protected readonly weekdayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  protected readonly monthDate = signal(startOfMonth(new Date()));
  protected readonly events = signal<EarningsEvent[]>([]);
  protected readonly isLoading = signal(false);
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly watchlistOnly = signal(true);
  protected readonly showPast = signal(false);
  protected readonly selectedDay = signal<string>(todayIso());
  protected readonly selectedEvent = signal<EarningsEvent | null>(null);

  protected readonly monthLabel = computed(() =>
    this.monthDate().toLocaleDateString('en-US', { month: 'long', year: 'numeric' }),
  );

  private readonly watchlistSymbols = computed(() => {
    const symbols = new Set(this.localWatchlist.symbols().map((s) => s.toUpperCase()));
    for (const item of this.watchlistApi.items()) {
      symbols.add(item.symbol.toUpperCase());
    }
    return symbols;
  });

  protected readonly watchlistEmpty = computed(() => this.watchlistSymbols().size === 0);

  private readonly filteredEvents = computed(() => {
    let items = this.events();
    if (this.watchlistOnly()) {
      const symbols = this.watchlistSymbols();
      items = items.filter((event) => symbols.has(event.symbol.toUpperCase()));
    }
    if (!this.showPast()) {
      const today = todayIso();
      items = items.filter((event) => earningsDateIso(event) >= today);
    }
    return items;
  });

  private readonly eventsByDate = computed(() => {
    const map = new Map<string, EarningsEvent[]>();
    for (const event of this.filteredEvents()) {
      const key = (event.date ?? '').slice(0, 10);
      if (!key) {
        continue;
      }
      const list = map.get(key);
      if (list) {
        list.push(event);
      } else {
        map.set(key, [event]);
      }
    }
    return map;
  });

  protected readonly days = computed<CalendarDay[]>(() => {
    const { gridStart, gridEnd } = gridRange(this.monthDate());
    const month = this.monthDate().getMonth();
    const today = todayIso();
    const byDate = this.eventsByDate();

    const days: CalendarDay[] = [];
    const cursor = new Date(gridStart);
    while (cursor <= gridEnd) {
      const iso = toIsoDate(cursor);
      const events = byDate.get(iso) ?? [];
      days.push({
        iso,
        dayNumber: cursor.getDate(),
        inMonth: cursor.getMonth() === month,
        isToday: iso === today,
        events,
        visibleEvents: events.slice(0, MAX_CHIPS_PER_DAY),
        moreCount: Math.max(0, events.length - MAX_CHIPS_PER_DAY),
      });
      cursor.setDate(cursor.getDate() + 1);
    }
    return days;
  });

  protected readonly monthEventCount = computed(() =>
    this.days().reduce((sum, day) => (day.inMonth ? sum + day.events.length : sum), 0),
  );

  protected readonly selectedDayEvents = computed(
    () => this.eventsByDate().get(this.selectedDay()) ?? [],
  );

  ngOnInit(): void {
    this.load();
    void this.watchlistApi.load().catch(() => undefined);
  }

  protected prevMonth(): void {
    this.shiftMonth(-1);
  }

  protected nextMonth(): void {
    this.shiftMonth(1);
  }

  protected goToday(): void {
    this.monthDate.set(startOfMonth(new Date()));
    this.selectedDay.set(todayIso());
    this.load();
  }

  protected selectDay(iso: string): void {
    this.selectedDay.set(iso);
  }

  protected openEventDialog(event: EarningsEvent, mouseEvent?: MouseEvent): void {
    mouseEvent?.stopPropagation();
    this.selectedEvent.set(event);
  }

  protected closeEventDialog(): void {
    this.selectedEvent.set(null);
  }

  @HostListener('document:keydown.escape')
  protected onEscape(): void {
    if (this.selectedEvent()) {
      this.closeEventDialog();
    }
  }

  protected toggleWatchlistOnly(): void {
    this.watchlistOnly.update((value) => !value);
  }

  protected toggleShowPast(): void {
    this.showPast.update((value) => !value);
    this.load();
  }

  protected isPastDay(iso: string): boolean {
    return isPastEarningsDate(iso);
  }

  protected hourLabel(hour: string | undefined | null): string {
    switch (hour) {
      case 'bmo':
        return 'Before open';
      case 'amc':
        return 'After close';
      case 'dmh':
        return 'During market';
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

  private shiftMonth(delta: number): void {
    const next = new Date(this.monthDate());
    next.setMonth(next.getMonth() + delta);
    this.monthDate.set(startOfMonth(next));
    this.load();
  }

  private load(): void {
    const { gridStart, gridEnd } = gridRange(this.monthDate());
    const from = toIsoDate(gridStart);
    const to = toIsoDate(gridEnd);
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.marketService.getEarnings(from, to).subscribe({
      next: (response) => {
        const calendarEvents = response.earnings ?? [];
        if (this.showPast()) {
          this.loadPastHistory(from, to, calendarEvents);
          return;
        }
        this.events.set(calendarEvents);
        this.isLoading.set(false);
      },
      error: (error: Error) => {
        this.events.set([]);
        this.errorMessage.set(error.message);
        this.isLoading.set(false);
      },
    });
  }

  private loadPastHistory(from: string, to: string, calendarEvents: EarningsEvent[]): void {
    const symbols = Array.from(this.watchlistSymbols());
    if (symbols.length === 0) {
      this.events.set(calendarEvents);
      this.isLoading.set(false);
      return;
    }

    this.marketService.getEarningsHistory(from, to, symbols).subscribe({
      next: (response) => {
        this.events.set(mergeEarningsEvents(calendarEvents, response.earnings ?? []));
        this.isLoading.set(false);
      },
      error: () => {
        this.events.set(calendarEvents);
        this.isLoading.set(false);
      },
    });
  }
}

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function gridRange(monthStart: Date): { gridStart: Date; gridEnd: Date } {
  const gridStart = new Date(monthStart);
  gridStart.setDate(gridStart.getDate() - gridStart.getDay());

  const monthEnd = new Date(monthStart.getFullYear(), monthStart.getMonth() + 1, 0);
  const gridEnd = new Date(monthEnd);
  gridEnd.setDate(gridEnd.getDate() + (6 - gridEnd.getDay()));

  return { gridStart, gridEnd };
}

function toIsoDate(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${month}-${day}`;
}
