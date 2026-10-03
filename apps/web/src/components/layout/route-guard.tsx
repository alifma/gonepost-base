'use client';

import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';

import { Icons } from '@/components/icons';
import { navGroups } from '@/config/nav-config';
import { useCan } from '@/lib/permissions';
import type { NavItem } from '@/types';

const items: NavItem[] = navGroups.flatMap((g) => g.items.flatMap((i) => [i, ...(i.items ?? [])]));

/** The permission a route needs, taken from the nav config so there is one source of truth. */
function requiredPermission(pathname: string) {
  const match = items
    .filter((i) => i.url !== '#' && (pathname === i.url || pathname.startsWith(`${i.url}/`)))
    .toSorted((a, b) => b.url.length - a.url.length)[0];
  return match?.access?.permission;
}

/**
 * Blocks a dashboard page when the user's roles don't grant what its nav
 * item requires. The API enforces the same rule; this only avoids a page
 * full of errors. Pages never check roles themselves.
 */
export function RouteGuard({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const { loaded, can } = useCan();
  const need = requiredPermission(pathname);
  if (!need) return children;
  if (!loaded) return null;
  if (can(need)) return children;
  return (
    <div className='flex flex-1 items-center justify-center p-6'>
      <div className='bg-card grid max-w-md gap-3 rounded-xl border p-6 text-center shadow-sm'>
        <span className='bg-muted text-muted-foreground mx-auto inline-flex size-10 items-center justify-center rounded-full'>
          <Icons.lock size={18} />
        </span>
        <h1 className='text-lg font-bold'>No access</h1>
        <p className='text-muted-foreground text-sm'>
          Your account has no role that grants <span className='text-foreground font-semibold'>{need}</span>. Ask an
          admin to assign one (Users page), then reload.
        </p>
      </div>
    </div>
  );
}
