import { Alert, Button, Paper, Stack, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import EntityTable from '../components/common/EntityTable';
import AssignRoleDialog from '../components/admin/AssignRoleDialog';
import DetailsDialog from '../components/common/DetailsDialog';
import { GridColDef, GridRenderCellParams } from '@mui/x-data-grid';

export function AdminPage() {
  const { me, token } = useAuth();
  const isAdmin = (me?.roles || []).includes('admin');
  const [users, setUsers] = useState<Array<Record<string, unknown>>>([]);
  const [assignOpen, setAssignOpen] = useState(false);
  const [assignTarget, setAssignTarget] = useState<number | null>(null);
  const [detail, setDetail] = useState<Record<string, unknown> | null>(null);

  async function load() {
    if (!token || !isAdmin) return;
    const out = await api.adminUsers(token);
    setUsers(out.items || []);
  }

  useEffect(() => { load().catch(() => {}); }, [token, isAdmin]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Admin Control Plane" subtitle="Role permissions and audit-sensitive actions." />
      {!isAdmin ? (
        <Alert severity="warning">Your account is not authorized for admin endpoints yet.</Alert>
      ) : (
        <>
          <Button variant="outlined" onClick={() => load()}>Refresh Users</Button>
          <Paper sx={{ p: 2.5 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Permission Governance</Typography>
            <Typography color="text.secondary" sx={{ mb: 1.5 }}>
              Role assignment endpoint is available at <code>/api/v1/admin/roles/assign</code> and each change is audited.
            </Typography>
            {users.length === 0 ? (
              <Alert severity="info">No users found.</Alert>
            ) : (
              <EntityTable
                rows={(users as any[]).map((u) => ({ id: Number(u.id), username: u.username, roles: Array.isArray(u.roles) ? (u.roles as any).join(', ') : '' }))}
                columns={[
                  { field: 'id', headerName: 'ID', width: 90 },
                  { field: 'username', headerName: 'Username', width: 240 },
                  { field: 'roles', headerName: 'Roles', width: 260 },
                  {
                    field: 'actions', headerName: 'Actions', width: 260, sortable: false, renderCell: (p: GridRenderCellParams) => (
                      <>
                        {isAdmin && <Button variant="outlined" size="small" onClick={() => { setAssignTarget(Number(p.row.id)); setAssignOpen(true); }} sx={{ mr: 1 }}>Assign Role</Button>}
                        <Button variant="text" size="small" onClick={async () => {
                          if (!token) return;
                          try {
                            const u = await api.getUser(token, Number(p.row.id));
                            setDetail(u);
                          } catch (e) {
                            setDetail({ error: (e as Error).message });
                          }
                        }}>View</Button>
                      </>
                    )
                  }
                ] as GridColDef[]}
              />
            )}
          </Paper>
          <AssignRoleDialog open={assignOpen} onClose={() => setAssignOpen(false)} targetUserId={assignTarget}
            onSubmit={async (role) => {
              if (!token || !assignTarget) return;
              await api.adminAssignRole(token, { targetUserId: assignTarget, role });
              await load();
            }}
          />
          <DetailsDialog open={!!detail} content={detail || {}} title="User Details" onClose={() => setDetail(null)} />
        </>
      )}
    </Stack>
  );
}
