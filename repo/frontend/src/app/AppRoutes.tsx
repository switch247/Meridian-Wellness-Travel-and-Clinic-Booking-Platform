import { Navigate, Route, Routes } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { RoleProtectedRoute } from './RoleProtectedRoute';
import { AppLayout } from '../layouts/AppLayout';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { CatalogPage } from '../pages/CatalogPage';
import { ProfilePage } from '../pages/ProfilePage';
import { AdminPage } from '../pages/AdminPage';
import { NotFoundPage } from '../pages/NotFoundPage';
import { MyReservationsPage } from '../pages/MyReservationsPage';
import { MyAgendaPage } from '../pages/MyAgendaPage';
import { AssignedSessionsPage } from '../pages/AssignedSessionsPage';
import { OpsSchedulingPage } from '../pages/OpsSchedulingPage';
import { RoleAuditPage } from '../pages/RoleAuditPage';
import { CommunityPage } from '../pages/CommunityPage';
import { NotificationsPage } from '../pages/NotificationsPage';
import { AnalyticsPage } from '../pages/AnalyticsPage';
import { EmailQueuePage } from '../pages/EmailQueuePage';

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/" element={<DashboardPage />} />

          <Route element={<RoleProtectedRoute roles={['traveler', 'operations', 'admin']} />}>
            <Route path="/catalog" element={<CatalogPage />} />
          </Route>
          <Route element={<RoleProtectedRoute roles={['traveler', 'coach', 'clinician', 'operations', 'admin']} />}>
            <Route path="/community" element={<CommunityPage />} />
            <Route path="/notifications" element={<NotificationsPage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['traveler', 'admin']} />}>
            <Route path="/profile" element={<ProfilePage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['traveler']} />}>
            <Route path="/my-reservations" element={<MyReservationsPage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['coach', 'clinician']} />}>
            <Route path="/my-agenda" element={<MyAgendaPage />} />
            <Route path="/assigned-sessions" element={<AssignedSessionsPage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['operations', 'admin']} />}>
            <Route path="/ops-scheduling" element={<OpsSchedulingPage />} />
            <Route path="/analytics" element={<AnalyticsPage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['operations', 'admin']} />}>
            <Route path="/email-queue" element={<EmailQueuePage />} />
          </Route>

          <Route element={<RoleProtectedRoute roles={['admin']} />}>
            <Route path="/role-audits" element={<RoleAuditPage />} />
            <Route path="/admin" element={<AdminPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="/404" element={<NotFoundPage />} />
      <Route path="*" element={<Navigate to="/404" replace />} />
    </Routes>
  );
}
