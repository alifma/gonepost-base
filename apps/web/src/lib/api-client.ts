import { createApiClient, type components } from '@gonepost/api-client';

const api = createApiClient(process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080');

export type CurrentUser = components['schemas']['User'];

export async function getCurrentUser(cookieHeader?: string): Promise<CurrentUser | null> {
  try {
    const { data, error } = await api.GET('/api/v1/auth/me', {
      credentials: 'include',
      headers: cookieHeader ? { cookie: cookieHeader } : undefined
    });

    return error ? null : data;
  } catch {
    return null;
  }
}

export { api };
