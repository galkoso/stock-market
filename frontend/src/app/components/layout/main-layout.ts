import { DatePipe } from '@angular/common';
import { Component, HostListener, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { NotificationRecord } from '../../models/market.model';
import { AuthService } from '../../services/auth.service';
import { MarketStreamCoordinatorService } from '../../services/market-stream-coordinator.service';
import { NotificationsService } from '../../services/notifications.service';

@Component({
  selector: 'app-main-layout',
  imports: [RouterOutlet, RouterLink, RouterLinkActive, DatePipe],
  templateUrl: './main-layout.html',
  styleUrl: './main-layout.scss',
})
export class MainLayout implements OnInit, OnDestroy {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly notificationsService = inject(NotificationsService);
  private readonly streamCoordinator = inject(MarketStreamCoordinatorService);

  protected readonly currentUser = this.authService.currentUser;
  protected readonly notifications = this.notificationsService.notifications;
  protected readonly unreadCount = this.notificationsService.unreadCount;
  protected readonly notificationsOpen = signal(false);

  ngOnInit(): void {
    void this.bootstrapNotifications();
  }

  ngOnDestroy(): void {
    this.notificationsService.disconnect();
    this.streamCoordinator.stop();
  }

  @HostListener('document:click')
  protected onDocumentClick(): void {
    this.notificationsOpen.set(false);
  }

  protected toggleNotifications(event: MouseEvent): void {
    event.stopPropagation();
    const next = !this.notificationsOpen();
    this.notificationsOpen.set(next);
    if (next) {
      void this.refreshNotifications();
    }
  }

  protected closeNotifications(): void {
    this.notificationsOpen.set(false);
  }

  protected async openNotification(item: NotificationRecord): Promise<void> {
    if (!item.isRead) {
      await this.notificationsService.markRead(item.id);
    }
    this.notificationsOpen.set(false);
    if (item.symbol) {
      await this.router.navigate(['/stock', item.symbol]);
    }
  }

  protected async markAllRead(): Promise<void> {
    await this.notificationsService.markAllRead();
  }

  protected async logout(): Promise<void> {
    this.notificationsService.disconnect();
    this.streamCoordinator.stop();
    await this.authService.logout();
    await this.router.navigate(['/login']);
  }

  private async bootstrapNotifications(): Promise<void> {
    try {
      await this.notificationsService.load();
      this.notificationsService.connect();
      await this.streamCoordinator.start();
    } catch {
      // Ignore initial load errors — streams reconnect when backend is available.
    }
  }

  private async refreshNotifications(): Promise<void> {
    try {
      await this.notificationsService.load();
    } catch {
      // Ignore refresh errors when opening the dropdown.
    }
  }
}
