import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const guestGuard: CanActivateFn = async () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.sessionVerified()) {
    await authService.bootstrap();
  }

  if (authService.currentUser()) {
    return router.createUrlTree(['/']);
  }

  return true;
};
