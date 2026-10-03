import { NavGroup } from '@/types';

/**
 * Navigation configuration with RBAC support
 *
 * This configuration is used for both the sidebar navigation and Cmd+K bar.
 * Items are organized into groups, each rendered with a SidebarGroupLabel.
 *
 * RBAC Access Control:
 * Each navigation item can have an `access` property that controls visibility
 * based on permissions, plans, features, roles, and organization context.
 *
 * Examples:
 *
 * 1. Require organization:
 *    access: { requireOrg: true }
 *
 * 2. Require specific permission:
 *    access: { requireOrg: true, permission: 'org:teams:manage' }
 *
 * 3. Require specific plan:
 *    access: { plan: 'pro' }
 *
 * 4. Require specific feature:
 *    access: { feature: 'premium_access' }
 *
 * 5. Require specific role:
 *    access: { role: 'admin' }
 *
 * 6. Multiple conditions (all must be true):
 *    access: { requireOrg: true, permission: 'org:teams:manage', plan: 'pro' }
 *
 * 7. Gate by RBAC permission (code from the API's permission list):
 *    access: { permission: 'items:read' }
 *    The sidebar and Cmd+K hide the item, and RouteGuard blocks the page
 *    (and sub-pages) from this same entry. Add one entry per page here.
 *
 * Note: The `visible` function is deprecated but still supported for backward compatibility.
 * Use the `access` property for new items.
 */
export const navGroups: NavGroup[] = [
  {
    label: 'Overview',
    items: [
      {
        title: 'Dashboard',
        url: '/dashboard/overview',
        icon: 'dashboard',
        isActive: false,
        shortcut: ['d', 'd'],
        items: []
      },
      {
        title: 'Items',
        url: '/dashboard/items',
        icon: 'product',
        shortcut: ['i', 'i'],
        isActive: false,
        items: [],
        access: { permission: 'items:read' }
      },
      {
        title: 'Users',
        url: '/dashboard/users',
        icon: 'teams',
        shortcut: ['u', 'u'],
        isActive: false,
        items: [],
        access: { permission: 'users:read' }
      },
      {
        title: 'Roles',
        url: '/dashboard/roles',
        icon: 'lock',
        shortcut: ['r', 'r'],
        isActive: false,
        items: [],
        access: { permission: 'roles:read' }
      },
      {
        title: 'Audit Logs',
        url: '/dashboard/audit-logs',
        icon: 'clock',
        shortcut: ['a', 'l'],
        isActive: false,
        items: [],
        access: { permission: 'audit:read' }
      }
    ]
  },
  {
    label: '',
    items: [
      {
        title: 'Account',
        url: '#',
        icon: 'account',
        isActive: true,
        items: [
          {
            title: 'Settings',
            shortcut: ['s', 's'],
            url: '/dashboard/settings',
            icon: 'settings'
          }
        ]
      }
    ]
  }
];
