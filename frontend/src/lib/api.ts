const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// Token Management di LocalStorage
export const getStoredToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('pmo_token');
};

export const setStoredToken = (token: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('pmo_token', token);
  }
};

export const clearStoredToken = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('pmo_token');
    localStorage.removeItem('pmo_user');
  }
};

// Generic Fetch Wrapper dengan Interceptor Token Otomatis
export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getStoredToken();
  const headers = new Headers(options.headers || {});

  // Sisipkan Authorization Header jika token ada
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  // Jangan set 'Content-Type': 'application/json' jika body adalah FormData (upload file)
  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
  });

  // Jika token kadaluarsa (HTTP 401), bersihkan storage
  if (response.status === 401) {
    clearStoredToken();
    if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
      window.location.href = '/login';
    }
  }

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const errorMsg = data.error || data.details || 'Terjadi kesalahan pada server';
    throw new Error(errorMsg);
  }

  return data as T;
}