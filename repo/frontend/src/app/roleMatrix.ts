export type Role = 'traveler' | 'coach' | 'clinician' | 'operations' | 'admin';

export type NavItem = {
  label: string;
  path: string;
  icon: string;
  roles: Role[];
};

export const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: 'dashboard', roles: ['traveler', 'coach', 'clinician', 'operations', 'admin'] },
  { label: 'Catalog', path: '/catalog', icon: 'catalog', roles: ['traveler', 'operations', 'admin'] },
  { label: 'Community', path: '/community', icon: 'community', roles: ['traveler', 'coach', 'clinician', 'operations', 'admin'] },
  { label: 'Notifications', path: '/notifications', icon: 'notifications', roles: ['traveler', 'coach', 'clinician', 'operations', 'admin'] },
  { label: 'Profile', path: '/profile', icon: 'profile', roles: ['traveler', 'admin'] },
  { label: 'My Reservations', path: '/my-reservations', icon: 'reservations', roles: ['traveler'] },
  { label: 'My Agenda', path: '/my-agenda', icon: 'agenda', roles: ['coach', 'clinician'] },
  { label: 'Assigned Sessions', path: '/assigned-sessions', icon: 'sessions', roles: ['coach', 'clinician'] },
  { label: 'Scheduling Ops', path: '/ops-scheduling', icon: 'ops', roles: ['operations', 'admin'] },
  { label: 'Analytics', path: '/analytics', icon: 'analytics', roles: ['operations', 'admin'] },
  { label: 'Email Queue', path: '/email-queue', icon: 'email', roles: ['operations', 'admin'] },
  { label: 'Role Audits', path: '/role-audits', icon: 'audits', roles: ['admin'] },
  { label: 'Admin', path: '/admin', icon: 'admin', roles: ['admin'] },
];

export function canAccess(roles: string[] | undefined, allowed: Role[]): boolean {
  const actual = roles || [];
  return actual.some((r) => allowed.includes(r as Role));
}
