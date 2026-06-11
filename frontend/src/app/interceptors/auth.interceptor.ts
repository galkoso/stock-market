import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, from, switchMap, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

const isAuthBootstrapRequest = (url: string): boolean =>
  url.includes('/api/auth/login') ||
  url.includes('/api/auth/register') ||
  url.includes('/api/auth/refresh');

const shouldTryRefresh = (error: HttpErrorResponse): boolean => {
  if (error.status !== 401) {
    return false;
  }

  const payload = error.error as { errorCode?: string } | null;
  const code = payload?.errorCode;
  return code === 'ACCESS_TOKEN_EXPIRED' || code === 'ACCESS_TOKEN_MISSING';
};

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);

  const withCredentials = req.clone({
    withCredentials: true,
    setHeaders: buildHeaders(req, authService),
  });

  return next(withCredentials).pipe(
    catchError((error: HttpErrorResponse) => {
      if (
        !(error instanceof HttpErrorResponse) ||
        isAuthBootstrapRequest(req.url) ||
        !shouldTryRefresh(error)
      ) {
        return throwError(() => error);
      }

      return from(authService.refreshAccessToken()).pipe(
        switchMap((user) => {
          if (!user) {
            return throwError(() => error);
          }

          const retryReq = req.clone({
            withCredentials: true,
            setHeaders: buildHeaders(req, authService),
          });

          return next(retryReq);
        }),
      );
    }),
  );
};

function buildHeaders(
  req: Parameters<HttpInterceptorFn>[0],
  authService: AuthService,
): Record<string, string> {
  const headers: Record<string, string> = {};
  const accessToken = authService.getStoredAccessToken();

  if (accessToken && !req.headers.has('Authorization')) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  return headers;
}
