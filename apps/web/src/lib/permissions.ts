'use client';

import { useQuery } from '@tanstack/react-query';

import { api } from '@/lib/api-client';
import { unwrap } from '@/lib/api-errors';

/**
 * Permission codes of the signed-in user (GET /auth/permissions). The API
 * is the source of truth and enforces every permission itself; the web only
 * uses this to hide menus, guard pages and hide write buttons.
 */
export function usePermissions() {
  return useQuery({
    queryKey: ['auth', 'permissions'],
    queryFn: () => unwrap(api.GET('/api/v1/auth/permissions', { credentials: 'include' })),
    staleTime: 5 * 60 * 1000
  });
}

/** `can('items:write')`. With no code it is always true; `loaded` is false until the list arrives. */
export function useCan() {
  const perms = usePermissions();
  const held = new Set(perms.data?.permissions ?? []);
  return { loaded: perms.isSuccess, can: (code?: string) => !code || held.has(code) };
}
