'use client';

import { useCan } from '@/lib/permissions';
import type { NavItem, NavGroup } from '@/types';

/**
 * Hides nav items whose `access.permission` the signed-in user does not
 * hold. While permissions load, gated items stay hidden.
 */
export function useFilteredNavItems(items: NavItem[]) {
  const { can } = useCan();
  return items.filter((item) => can(item.access?.permission));
}

export function useFilteredNavGroups(groups: NavGroup[]) {
  const { can } = useCan();
  return groups
    .map((group) => ({ ...group, items: group.items.filter((item) => can(item.access?.permission)) }))
    .filter((group) => group.items.length > 0);
}
