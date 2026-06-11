export interface User {
  id: string;
  username: string;
}

export interface AuthResponse {
  success?: boolean;
  user?: User;
  accessToken?: string;
  error?: string;
  errorCode?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}
