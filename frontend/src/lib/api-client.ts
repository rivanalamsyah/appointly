import ky from 'ky';
import type { ApiResponse } from '@/types/api';

/**
 * Global API Client configured for Appointly SaaS Backend.
 * Automatically injects Auth Bearer Token and Tenant Header.
 */

const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('appointly_auth_token');
};

const getTenantSlug = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('appointly_tenant_slug');
};

export const apiClient = ky.create({
  prefixUrl: typeof window !== 'undefined' ? '' : (process.env.API_URL || 'http://localhost:8080'),
  timeout: 15000,
  hooks: {
    beforeRequest: [
      (request: Request) => {
        const token = getAuthToken();
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
        const tenantSlug = getTenantSlug();
        if (tenantSlug) {
          request.headers.set('X-Tenant-Slug', tenantSlug);
        }
      },
    ],
    afterResponse: [
      async (_request: Request, _options: unknown, response: Response) => {
        if (response.status === 401 && typeof window !== 'undefined') {
          // Token expired or invalid
          localStorage.removeItem('appointly_auth_token');
          if (!window.location.pathname.startsWith('/auth/login')) {
            window.location.href = '/auth/login?reason=expired';
          }
        }
      },
    ],
  },
});

/**
 * Helper wrapper for API requests that unwraps standard API envelope response.
 */
export async function fetchApi<T>(url: string, options?: Parameters<typeof apiClient>[1]): Promise<T> {
  const res = await apiClient(url, options).json<ApiResponse<T>>();
  if (!res.success) {
    throw new Error(res.error?.message || 'API Request failed');
  }
  return res.data as T;
}
