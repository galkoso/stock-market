import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  imports: [FormsModule],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class LoginPage {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly mode = signal<'signin' | 'signup'>('signin');
  protected readonly username = signal('');
  protected readonly password = signal('');
  protected readonly errorMessage = signal<string | null>(null);
  protected readonly isSubmitting = signal(false);

  protected setMode(mode: 'signin' | 'signup'): void {
    this.mode.set(mode);
    this.errorMessage.set(null);
  }

  protected async onSubmit(event: Event): Promise<void> {
    event.preventDefault();
    if (this.isSubmitting()) {
      return;
    }

    const username = this.username().trim();
    const password = this.password().trim();

    if (!username || !password) {
      this.errorMessage.set('Username and password are required.');
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);

    const error =
      this.mode() === 'signin'
        ? await this.authService.login({ username, password })
        : await this.authService.register({ username, password });

    this.isSubmitting.set(false);

    if (error) {
      this.errorMessage.set(error);
      return;
    }

    await this.router.navigateByUrl('/');
  }
}
