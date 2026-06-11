import { DatePipe } from '@angular/common';
import { Component, HostListener, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { NotificationRecord } from '../../models/market.model';
import { AuthService } from '../../services/auth.service';
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

  private pollTimer: ReturnType<typeof setInterval> | null = null;

  protected readonly currentUser = this.authService.currentUser;
  protected readonly notifications = this.notificationsService.notifications;
  protected readonly unreadCount = this.notificationsService.unreadCount;
  protected readonly notificationsOpen = signal(false);

  ngOnInit(): void {
    void this.refreshNotifications();
    this.pollTimer = setInterval(() => {
      void this.refreshNotifications();
    }, 30_000);
  }

  ngOnDestroy(): void {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
    }
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
    await this.authService.logout();
    await this.router.navigate(['/login']);
  }

  private async refreshNotifications(): Promise<void> {
    try {
      await this.notificationsService.load();
    } catch {
      // Ignore polling errors — user may be offline briefly.
    }
  }
}
