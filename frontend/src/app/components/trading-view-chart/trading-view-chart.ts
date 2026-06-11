import {
  AfterViewInit,
  Component,
  effect,
  ElementRef,
  input,
  OnDestroy,
  signal,
  viewChild,
} from '@angular/core';

const WIDGET_SCRIPT =
  'https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js';
const STORAGE_KEY = 'stock-market-chart-height';
const DEFAULT_HEIGHT = 640;
const MIN_HEIGHT = 360;

@Component({
  selector: 'app-trading-view-chart',
  templateUrl: './trading-view-chart.html',
  styleUrl: './trading-view-chart.scss',
})
export class TradingViewChart implements AfterViewInit, OnDestroy {
  readonly symbol = input.required<string>();

  protected readonly chartHeight = signal(loadStoredHeight());

  private readonly shell = viewChild.required<ElementRef<HTMLDivElement>>('shell');
  private readonly container = viewChild.required<ElementRef<HTMLDivElement>>('container');
  private readonly viewReady = signal(false);
  private resizeObserver: ResizeObserver | null = null;

  constructor() {
    effect(() => {
      const symbol = this.symbol();
      if (!this.viewReady()) {
        return;
      }

      this.renderWidget(symbol);
    });
  }

  ngAfterViewInit(): void {
    this.viewReady.set(true);
    this.resizeObserver = new ResizeObserver(() => {
      window.dispatchEvent(new Event('resize'));
    });
    this.resizeObserver.observe(this.shell().nativeElement);
  }

  ngOnDestroy(): void {
    this.resizeObserver?.disconnect();
  }

  protected startResize(event: MouseEvent | TouchEvent): void {
    event.preventDefault();

    const startY = getPointerY(event);
    const startHeight = this.chartHeight();

    const onMove = (moveEvent: MouseEvent | TouchEvent) => {
      const delta = getPointerY(moveEvent) - startY;
      this.chartHeight.set(clampHeight(startHeight + delta));
    };

    const onEnd = () => {
      localStorage.setItem(STORAGE_KEY, String(this.chartHeight()));
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onEnd);
      document.removeEventListener('touchmove', onMove);
      document.removeEventListener('touchend', onEnd);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      window.dispatchEvent(new Event('resize'));
    };

    document.body.style.cursor = 'ns-resize';
    document.body.style.userSelect = 'none';
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onEnd);
    document.addEventListener('touchmove', onMove, { passive: false });
    document.addEventListener('touchend', onEnd);
  }

  private renderWidget(symbol: string): void {
    const host = this.container().nativeElement;
    host.innerHTML = '';

    const widget = document.createElement('div');
    widget.className = 'tradingview-widget-container__widget';
    widget.style.height = '100%';
    widget.style.width = '100%';

    const script = document.createElement('script');
    script.type = 'text/javascript';
    script.src = WIDGET_SCRIPT;
    script.async = true;
    script.innerHTML = JSON.stringify({
      autosize: true,
      symbol: toTradingViewSymbol(symbol),
      interval: 'D',
      timezone: 'exchange',
      theme: 'dark',
      style: '1',
      locale: 'en',
      backgroundColor: 'rgba(13, 20, 36, 1)',
      gridColor: 'rgba(51, 65, 85, 0.35)',
      hide_side_toolbar: false,
      allow_symbol_change: false,
      calendar: false,
      support_host: 'https://www.tradingview.com',
      withdateranges: true,
      save_image: false,
    });

    host.appendChild(widget);
    host.appendChild(script);
  }
}

function toTradingViewSymbol(symbol: string): string {
  return `NASDAQ:${symbol}`;
}

function loadStoredHeight(): number {
  const stored = Number(localStorage.getItem(STORAGE_KEY));
  if (!Number.isFinite(stored)) {
    return DEFAULT_HEIGHT;
  }

  return clampHeight(stored);
}

function clampHeight(height: number): number {
  const maxHeight = Math.max(MIN_HEIGHT, Math.floor(window.innerHeight * 0.85));
  return Math.min(maxHeight, Math.max(MIN_HEIGHT, Math.round(height)));
}

function getPointerY(event: MouseEvent | TouchEvent): number {
  if ('touches' in event) {
    return event.touches[0]?.clientY ?? event.changedTouches[0]?.clientY ?? 0;
  }

  return event.clientY;
}
