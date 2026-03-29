import { Alert, Button, Paper, Stack, Typography, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import EntityTable from '../components/common/EntityTable';
import { GridColDef, GridRenderCellParams } from '@mui/x-data-grid';

export function RoleAuditPage() {
  const { token } = useAuth();
    const [items, setItems] = useState<Array<Record<string, unknown>>>([]);
    const [detail, setDetail] = useState<Record<string, unknown> | null>(null);

  async function load() {
    if (!token) return;
    const out = await api.adminRoleAudits(token);
    setItems(out.items || []);
  }

  useEffect(() => { load().catch(() => {}); }, [token]);

    return (
      <Stack spacing={2.5}>
        <SectionHeader title="Role Audits" subtitle="Permission mutation trace for governance." />
        <Button variant="outlined" onClick={() => load()}>Refresh</Button>
        <Paper sx={{ p: 2.5 }}>
          {items.length === 0 ? (
            <Alert severity="info">No role changes yet.</Alert>
          ) : (
            <EntityTable
              rows={(items as any[]).map((it) => ({ id: Number(it.id), actorId: it.actorId, targetUserId: it.targetUserId, action: it.action, createdAt: it.createdAt }))}
              columns={[
                { field: 'id', headerName: 'ID', width: 90 },
                { field: 'actorId', headerName: 'Actor', width: 140 },
                { field: 'targetUserId', headerName: 'Target', width: 140 },
                { field: 'action', headerName: 'Action', width: 200 },
                { field: 'createdAt', headerName: 'When', width: 200 },
                { field: 'actions', headerName: 'Actions', width: 140, sortable: false, renderCell: (p: GridRenderCellParams) => (
                    <Button size="small" variant="outlined" onClick={() => setDetail(p.row)}>Details</Button>
                  )}
              ] as GridColDef[]}
            />
          )}
        </Paper>

        <Dialog open={!!detail} onClose={() => setDetail(null)} fullWidth>
          <DialogTitle>Audit Details</DialogTitle>
          <DialogContent>
            <Typography variant="body2">Before:</Typography>
            <Paper sx={{ p: 1, my: 1, whiteSpace: 'pre-wrap' }}>{String(detail?.before || '')}</Paper>
            <Typography variant="body2">After:</Typography>
            <Paper sx={{ p: 1, my: 1, whiteSpace: 'pre-wrap' }}>{String(detail?.after || '')}</Paper>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDetail(null)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Stack>
    );
}
