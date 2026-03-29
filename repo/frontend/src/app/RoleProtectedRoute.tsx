import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export function RoleProtectedRoute({ roles }: { roles: string[] }) {
  const { me } = useAuth();
  const hasRole = (me?.roles || []).some((r) => roles.includes(r));
  if (!hasRole) {
    return <Navigate to="/" replace />;
  }
  return <Outlet />;
}
