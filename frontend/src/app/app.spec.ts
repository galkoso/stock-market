import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { App } from './app';
import { StockStreamService } from './services/stock-stream.service';

describe('App', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [provideHttpClient(), StockStreamService],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(App);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });
});
