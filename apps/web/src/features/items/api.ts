'use client';

import type { components } from '@gonepost/api-client';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { api } from '@/lib/api-client';
import { unwrap } from '@/lib/api-errors';

export type Item = components['schemas']['Item'];
export type ItemInput = components['schemas']['ItemInput'];
export type ItemStatus = components['schemas']['ItemStatus'];

const opts = { credentials: 'include' as const };

/** Query-key factory: every items query starts with ['items'], so one invalidate refreshes them all. */
export const itemKeys = {
  all: ['items'] as const,
  list: (query: { search?: string; status?: ItemStatus; limit: number; offset: number }) =>
    [...itemKeys.all, 'list', query] as const
};

export function useItems(query: { search?: string; status?: ItemStatus; limit: number; offset: number }) {
  return useQuery({
    queryKey: itemKeys.list(query),
    queryFn: () => unwrap(api.GET('/api/v1/items', { ...opts, params: { query } })),
    placeholderData: keepPreviousData
  });
}

/** Shared mutation plumbing: toast on result, refresh every items query on success. */
function useItemMutation<V>(fn: (vars: V) => Promise<unknown>, successMessage: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: fn,
    onSuccess: () => {
      toast.success(successMessage);
      return qc.invalidateQueries({ queryKey: itemKeys.all });
    },
    onError: (err: Error) => toast.error(err.message)
  });
}

export function useCreateItem() {
  return useItemMutation(
    (body: ItemInput) => unwrap(api.POST('/api/v1/items', { ...opts, body })),
    'Item created'
  );
}

export function useUpdateItem() {
  return useItemMutation(
    ({ id, body }: { id: string; body: ItemInput }) =>
      unwrap(api.PATCH('/api/v1/items/{id}', { ...opts, params: { path: { id } }, body })),
    'Item updated'
  );
}

export function useDeleteItem() {
  return useItemMutation(
    (id: string) => unwrap(api.DELETE('/api/v1/items/{id}', { ...opts, params: { path: { id } } })),
    'Item deleted'
  );
}
