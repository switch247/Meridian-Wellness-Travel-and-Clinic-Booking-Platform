const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8443/api/v1';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    }
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error((body as { error?: string }).error || 'Request failed');
  }
  return body as T;
}

export type LoginResult = { token: string; roles: string[] };
export type MeResult = { id: number; username: string; roles: string[]; phone: string; address: string };

export const api = {
  register: (payload: { username: string; password: string; phone: string; address: string }) =>
    request<{ id: number }>('/auth/register', { method: 'POST', body: JSON.stringify(payload) }),
  login: (payload: { username: string; password: string }) =>
    request<LoginResult>('/auth/login', { method: 'POST', body: JSON.stringify(payload) }),
  me: (token: string) =>
    request<MeResult>('/auth/me', { headers: { Authorization: `Bearer ${token}` } }),
  catalog: () => request<{ items: Array<Record<string, unknown>> }>('/catalog'),
  addAddress: (
    token: string,
    payload: { line1: string; line2: string; city: string; state: string; postalCode: string }
  ) => request('/profile/addresses', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  listAddresses: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/profile/addresses', { headers: { Authorization: `Bearer ${token}` } }),
  listContacts: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/profile/contacts', { headers: { Authorization: `Bearer ${token}` } }),
  addContact: (token: string, payload: { name: string; relationship?: string; phone: string }) =>
    request('/profile/contacts', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  deleteContact: (token: string, id: number) => request(`/profile/contacts/${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } }),
  placeHold: (
    token: string,
    payload: { packageId: number; hostId: number; roomId: number; slotStart: string; duration: number }
  ) => request('/bookings/holds', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  listHolds: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/bookings/holds', { headers: { Authorization: `Bearer ${token}` } }),
  listHistory: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/bookings/history', { headers: { Authorization: `Bearer ${token}` } }),
  adminUsers: (token: string, role?: string) => request<{ items: Array<Record<string, unknown>> }>(`/admin/users${role ? `?role=${encodeURIComponent(role)}` : ''}`, { headers: { Authorization: `Bearer ${token}` } }),
  adminRoleAudits: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/admin/roles/audits', { headers: { Authorization: `Bearer ${token}` } }),
  adminAssignRole: (token: string, payload: { targetUserId: number; role: string }) => request('/admin/roles/assign', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  getUser: (token: string, id: number) => request<{ id: number; username: string }>(`/users/${id}`, { headers: { Authorization: `Bearer ${token}` } }),
  deleteAddress: (token: string, id: number) => request(`/profile/addresses/${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } }),
  cancelHold: (token: string, id: number) => request(`/bookings/holds/${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } }),
  hostAgenda: (token: string, hostId: number) => request<{ items: Array<Record<string, unknown>> }>(`/scheduling/hosts/${hostId}/agenda`, { headers: { Authorization: `Bearer ${token}` } }),
  roomAgenda: (token: string, roomId: number) => request<{ items: Array<Record<string, unknown>> }>(`/scheduling/rooms/${roomId}/agenda`, { headers: { Authorization: `Bearer ${token}` } })
  ,
  routes: () => request<{ items: Array<Record<string, unknown>> }>('/catalog/routes'),
  hotels: () => request<{ items: Array<Record<string, unknown>> }>('/catalog/hotels'),
  attractions: () => request<{ items: Array<Record<string, unknown>> }>('/catalog/attractions'),
  availableSlots: (token: string, q: { hostId: number; roomId: number; day: string; duration: number }) =>
    request<{ items: Array<Record<string, unknown>> }>(`/scheduling/slots?hostId=${q.hostId}&roomId=${q.roomId}&day=${encodeURIComponent(q.day)}&duration=${q.duration}`, { headers: { Authorization: `Bearer ${token}` } }),
  listHosts: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/scheduling/hosts', { headers: { Authorization: `Bearer ${token}` } }),
  confirmHold: (token: string, payload: { holdId: number; version?: number }) =>
    request('/bookings/confirm', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  communityPosts: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/community/posts', { headers: { Authorization: `Bearer ${token}` } }),
  createCommunityPost: (token: string, payload: { title: string; body: string; destinationId?: number }) =>
    request('/community/posts', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  communityComments: (token: string, postId: number) => request<{ items: Array<Record<string, unknown>> }>(`/community/posts/${postId}/comments`, { headers: { Authorization: `Bearer ${token}` } }),
  addComment: (token: string, postId: number, payload: { body: string; parentCommentId?: number }) =>
    request(`/community/posts/${postId}/comments`, { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  likeTarget: (token: string, payload: { targetType: 'post' | 'comment'; targetId: number }) =>
    request('/community/likes', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  favoritePackage: (token: string, payload: { packageId: number }) =>
    request('/community/favorites', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  followUser: (token: string, payload: { userId: number }) =>
    request('/community/follows', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  blockUser: (token: string, payload: { userId: number }) =>
    request('/community/blocks', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  reportContent: (token: string, payload: { targetType: string; targetId: number; reason: string }) =>
    request('/community/reports', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  notifications: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/notifications', { headers: { Authorization: `Bearer ${token}` } }),
  markNotificationRead: (token: string, id: number) => request(`/notifications/${id}/read`, { method: 'POST', headers: { Authorization: `Bearer ${token}` } }),
  analyticsKpis: (token: string, q: { from: string; to: string; providerId?: number; packageId?: number }) =>
    request<{ kpis: Record<string, unknown> }>(`/ops/analytics/kpis?from=${q.from}&to=${q.to}${q.providerId ? `&providerId=${q.providerId}` : ''}${q.packageId ? `&packageId=${q.packageId}` : ''}`, { headers: { Authorization: `Bearer ${token}` } }),
  exportAnalytics: (token: string, q: { from: string; to: string; providerId?: number; packageId?: number }) =>
    request<{ path: string }>(`/ops/analytics/export?from=${q.from}&to=${q.to}${q.providerId ? `&providerId=${q.providerId}` : ''}${q.packageId ? `&packageId=${q.packageId}` : ''}`, { headers: { Authorization: `Bearer ${token}` } }),
  scheduleReport: (token: string, payload: { reportType: string; parameters: Record<string, unknown>; scheduledFor: string }) =>
    request('/ops/reports/schedule', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  queueEmail: (token: string, payload: { templateKey: string; recipientLabel: string; subject: string; body: string }) =>
    request('/ops/email/queue', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }),
  emailQueue: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/ops/email/queue', { headers: { Authorization: `Bearer ${token}` } }),
  exportEmailQueue: (token: string) => request<{ path: string }>('/ops/email/export', { method: 'POST', headers: { Authorization: `Bearer ${token}` } }),
  reportJobs: (token: string) => request<{ items: Array<Record<string, unknown>> }>('/ops/reports', { headers: { Authorization: `Bearer ${token}` } })
};
