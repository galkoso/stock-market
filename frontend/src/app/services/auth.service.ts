import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { normalizeUser } from '../lib/auth-response';
import { AuthResponse, LoginRequest, RegisterRequest, User } from '../models/user.model';

const AUTH_USER_STORAGE_KEY = 'stock_market_current_user';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);

  private accessToken: string | null = null;

  readonly currentUser = signal<User | null | undefined>(this.readStoredUser() ?? undefined);
  readonly sessionVerified = signal(false);

  getStoredAccessToken(): string | null {
    return this.accessToken;
  }

  async bootstrap(): Promise<void> {
    if (this.sessionVerified()) {
      return;
    }

    try {
      const user = await this.refreshAccessToken();
      this.currentUser.set(user);
    } catch {
      this.clearSession();
      this.currentUser.set(null);
    } finally {
      this.sessionVerified.set(true);
    }
  }

  async login(credentials: LoginRequest): Promise<string | null> {
    try {
      const response = await firstValueFrom(
        this.http.post<AuthResponse>('/api/auth/login', credentials, {
          withCredentials: true,
        }),
      );

      if (!response.success || !response.user || !response.accessToken) {
        return response.error ?? 'Invalid username or password';
      }

      this.setSession(response.accessToken, response.user);
      return null;
    } catch (error) {
      return this.readHttpError(error, 'Invalid username or password');
    }
  }

  async register(credentials: RegisterRequest): Promise<string | null> {
    try {
      const response = await firstValueFrom(
        this.http.post<AuthResponse>('/api/auth/register', credentials, {
          withCredentials: true,
        }),
      );

      if (!response.success || !response.user || !response.accessToken) {
        return response.error ?? 'Registration failed';
      }

      this.setSession(response.accessToken, response.user);
      return null;
    } catch (error) {
      return this.readHttpError(error, 'Registration failed');
    }
  }

  async refreshAccessToken(): Promise<User | null> {
    try {
      const response = await firstValueFrom(
        this.http.get<AuthResponse>('/api/auth/refresh', {
          withCredentials: true,
        }),
      );

      if (!response.success || !response.accessToken || !response.user) {
        this.clearSession();
        return null;
      }

      this.setSession(response.accessToken, response.user);
      return response.user;
    } catch {
      this.clearSession();
      return null;
    }
  }

  async logout(): Promise<void> {
    try {
      await firstValueFrom(
        this.http.post('/api/auth/logout', {}, { withCredentials: true }),
      );
    } finally {
      this.clearSession();
      this.currentUser.set(null);
      this.sessionVerified.set(true);
    }
  }

  private setSession(accessToken: string, user: User): void {
    this.accessToken = accessToken;
    this.persistUser(user);
    this.currentUser.set(user);
    this.sessionVerified.set(true);
  }

  private clearSession(): void {
    this.accessToken = null;
    this.persistUser(null);
  }

  private persistUser(user: User | null): void {
    if (!user) {
      localStorage.removeItem(AUTH_USER_STORAGE_KEY);
      return;
    }

    localStorage.setItem(AUTH_USER_STORAGE_KEY, JSON.stringify(user));
  }

  private readHttpError(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      const payload = error.error as AuthResponse | null;
      return payload?.error ?? fallback;
    }

    return 'Server connection failed';
  }

  readStoredUser(): User | null {
    const raw = localStorage.getItem(AUTH_USER_STORAGE_KEY);
    if (!raw) {
      return null;
    }

    try {
      return normalizeUser(JSON.parse(raw));
    } catch {
      localStorage.removeItem(AUTH_USER_STORAGE_KEY);
      return null;
    }
  }
}
