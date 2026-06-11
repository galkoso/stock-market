import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { SearchResult } from '../../models/market.model';
import { MarketService } from '../../services/market.service';

@Component({
  selector: 'app-search-page',
  imports: [FormsModule, RouterLink],
  templateUrl: './search.html',
  styleUrl: './search.scss',
})
export class SearchPage {
  private readonly marketService = inject(MarketService);

  protected readonly query = signal('');
  protected readonly results = signal<SearchResult[]>([]);
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly isLoading = signal(false);

  protected onSearch(event?: Event): void {
    event?.preventDefault();
    const q = this.query().trim();
    if (!q) return;

    this.isLoading.set(true);
    this.errorMessage.set(null);
    this.marketService.search(q).subscribe({
      next: (response) => {
        this.results.set(response.results ?? []);
        this.isLoading.set(false);
      },
      error: (error: Error) => {
        this.errorMessage.set(error.message);
        this.results.set([]);
        this.isLoading.set(false);
      },
    });
  }
}
