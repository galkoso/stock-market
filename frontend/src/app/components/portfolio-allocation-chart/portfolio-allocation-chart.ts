import { Component, computed, input } from '@angular/core';
import { NgxEchartsDirective } from 'ngx-echarts';
import type { EChartsCoreOption } from 'echarts/core';
import { AllocationHoldingRecord, LiveAllocationHoldingRecord } from '../../models/market.model';

const SLICE_COLORS = [
  '#818cf8',
  '#34d399',
  '#60a5fa',
  '#fbbf24',
  '#fb7185',
  '#a78bfa',
  '#38bdf8',
  '#4ade80',
  '#f472b6',
  '#fcd34d',
];

@Component({
  selector: 'app-portfolio-allocation-chart',
  imports: [NgxEchartsDirective],
  templateUrl: './portfolio-allocation-chart.html',
  styleUrl: './portfolio-allocation-chart.scss',
})
export class PortfolioAllocationChart {
  readonly holdings = input.required<AllocationHoldingRecord[] | LiveAllocationHoldingRecord[]>();
  readonly totalValue = input.required<number>();
  readonly totalValueIls = input.required<number>();

  protected readonly chartOptions = computed<EChartsCoreOption>(() => {
    const data = this.holdings();
    const total = this.totalValue();

    if (data.length === 0 || total <= 0) {
      return {
        backgroundColor: 'transparent',
        title: {
          text: 'No allocation data',
          left: 'center',
          top: 'center',
          textStyle: {
            color: '#6b7a94',
            fontSize: 14,
            fontWeight: 500,
          },
        },
      };
    }

    return {
      backgroundColor: 'transparent',
      color: SLICE_COLORS,
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(16, 24, 40, 0.95)',
        borderColor: 'rgba(148, 163, 184, 0.2)',
        textStyle: { color: '#e8edf7' },
        formatter: (params: {
          name: string;
          percent: number;
          data: { marketValue: number; marketValueIls: number };
        }) =>
          `<strong>${params.name}</strong><br/>` +
          `${params.percent.toFixed(2)}%<br/>` +
          `$${params.data.marketValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}<br/>` +
          `₪${params.data.marketValueIls.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      },
      legend: {
        orient: 'vertical',
        right: 0,
        top: 'center',
        textStyle: { color: '#aab8d0', fontSize: 12 },
        formatter: (name: string) => {
          const item = data.find((holding) => holding.symbol === name);
          if (!item) {
            return name;
          }
          return `${name}  ${item.allocationPercent.toFixed(1)}%`;
        },
      },
      series: [
        {
          name: 'Allocation',
          type: 'pie',
          radius: ['42%', '72%'],
          center: ['38%', '50%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 6,
            borderColor: '#070b14',
            borderWidth: 2,
          },
          label: {
            show: true,
            color: '#e8edf7',
            formatter: (params: { name: string; percent: number }) =>
              `${params.name}\n${params.percent.toFixed(1)}%`,
            fontSize: 11,
            fontWeight: 600,
          },
          labelLine: {
            lineStyle: { color: 'rgba(148, 163, 184, 0.35)' },
          },
          emphasis: {
            scaleSize: 8,
            label: { fontSize: 13, fontWeight: 700 },
          },
          data: data.map((holding) => ({
            name: holding.symbol,
            value: holding.marketValue,
            marketValue: holding.marketValue,
            marketValueIls: holding.marketValueIls,
          })),
        },
      ],
    };
  });
}
