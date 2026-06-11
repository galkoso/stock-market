import { Routes } from '@angular/router';
import { MainLayout } from './components/layout/main-layout';
import { authGuard } from './guards/auth.guard';
import { guestGuard } from './guards/guest.guard';
import { LoginPage } from './pages/login/login';

export const routes: Routes = [
  {
    path: 'login',
    component: LoginPage,
    canActivate: [guestGuard],
  },
  {
    path: '',
    component: MainLayout,
    canActivate: [authGuard],
    children: [
      {
        path: '',
        loadComponent: () => import('./app').then((m) => m.App),
      },
      {
        path: 'search',
        loadComponent: () => import('./pages/search/search').then((m) => m.SearchPage),
      },
      {
        path: 'stock/:symbol',
        loadComponent: () => import('./pages/stock-details/stock-details').then((m) => m.StockDetailsPage),
      },
      {
        path: 'earnings',
        loadComponent: () => import('./pages/earnings/earnings').then((m) => m.EarningsPage),
      },
      {
        path: 'calendar',
        loadComponent: () => import('./pages/calendar/calendar-page').then((m) => m.CalendarPage),
      },
      {
        path: 'watchlist',
        loadComponent: () => import('./pages/watchlist-page/watchlist-page').then((m) => m.WatchlistPage),
      },
      {
        path: 'alerts',
        loadComponent: () => import('./pages/alerts-page/alerts-page').then((m) => m.AlertsPage),
      },
    ],
  },
  {
    path: '**',
    redirectTo: '',
  },
];
