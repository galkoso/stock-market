import { AuthResponse, User } from '../models/user.model';

export const normalizeUser = (value: unknown): User | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const user = value as Record<string, unknown>;
  if (typeof user['id'] !== 'string' || typeof user['username'] !== 'string') {
    return null;
  }

  return {
    id: user['id'],
    username: user['username'],
  };
};

export const readAuthResponse = async (response: Response): Promise<AuthResponse> => {
  const payload: unknown = await response.json().catch(() => ({}));
  if (typeof payload !== 'object' || payload === null) {
    return {};
  }

  const user = 'user' in payload ? normalizeUser(payload.user) : null;

  return {
    success: 'success' in payload && payload.success === true,
    user: user ?? undefined,
    accessToken:
      'accessToken' in payload && typeof payload.accessToken === 'string'
        ? payload.accessToken
        : undefined,
    error: 'error' in payload && typeof payload.error === 'string' ? payload.error : undefined,
    errorCode:
      'errorCode' in payload && typeof payload.errorCode === 'string'
        ? payload.errorCode
        : undefined,
  };
};
