import { Alert, Button, Paper, Stack, Typography, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import DetailsDialog from '../components/common/DetailsDialog';
import EntityTable from '../components/common/EntityTable';
import { GridColDef } from '@mui/x-data-grid';

export function MyReservationsPage() {
  const { token } = useAuth();
  const [holds, setHolds] = useState<Array<Record<string, unknown>>>([]);
  const [history, setHistory] = useState<Array<Record<string, unknown>>>([]);

  async function load() {
    if (!token) return;
    const h = await api.listHolds(token);
    const his = await api.listHistory(token);
    setHolds(h.items || []);
    setHistory(his.items || []);
  }

  useEffect(() => { load().catch(() => {}); }, [token]);
  const [detail, setDetail] = useState<Record<string, unknown> | null>(null);
  const [confirm, setConfirm] = useState<{ open: boolean; id?: number | null; message?: string }>({ open: false });

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="My Reservations" subtitle="Active holds and historical reservations." />
      <Button variant="outlined" onClick={() => load()}>Refresh</Button>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Active Holds</Typography>
        {holds.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1 }}>No reservations yet. Start from Catalog or Booking.</Alert>
        ) : (
          <EntityTable
            rows={(holds as any[]).map(h => ({ id: Number(h.id), packageId: h.packageId, status: h.status, slotStart: h.slotStart, _raw: h }))}
            columns={[{ field: 'id', headerName: 'ID', width: 90 }, { field: 'packageId', headerName: 'Package', width: 140 }, { field: 'slotStart', headerName: 'Start', width: 220 }, { field: 'status', headerName: 'Status', width: 120 }, { field: 'actions', headerName: 'Actions', width: 240, sortable: false, renderCell: (p) => (
              <Stack direction="row" spacing={1}>
                <Button size="small" variant="outlined" onClick={() => setDetail(p.row._raw)}>Details</Button>
                <Button size="small" variant="contained" onClick={async () => {
                  if (!token) return;
                  await api.confirmHold(token, { holdId: Number(p.row.id), version: Number(p.row._raw?.version ?? 0) });
                  await load();
                }}>Confirm</Button>
                <Button size="small" color="error" variant="outlined" onClick={() => setConfirm({ open: true, id: Number(p.row.id), message: 'Cancel this hold?' })}>Cancel</Button>
              </Stack>
            ) } ] as GridColDef[]}
            height={280}
          />
        )}
      </Paper>
      <Dialog open={confirm.open} onClose={() => setConfirm({ open: false })}>
        <DialogTitle>Confirm</DialogTitle>
        <DialogContent>
          <Typography>{confirm.message}</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirm({ open: false })}>Cancel</Button>
          <Button color="error" onClick={async () => {
            try {
              if (!token) throw new Error('Please login');
              const id = Number(confirm.id);
              // optimistic UI: remove hold immediately
              setHolds((prev) => prev.filter((h: any) => Number(h.id) !== id));
              await api.cancelHold(token, id);
            } catch (err) {
              // on error, reload to restore state
              await load();
            } finally {
              setConfirm({ open: false });
            }
          }}>Cancel Hold</Button>
        </DialogActions>
      </Dialog>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">History</Typography>
        {history.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1 }}>No booking history yet.</Alert>
        ) : (
          <EntityTable
            rows={(history as any[]).map(h => ({ id: Number(h.id), packageId: h.packageId, status: h.status, slotStart: h.slotStart }))}
            columns={[{ field: 'id', headerName: 'ID', width: 90 }, { field: 'packageId', headerName: 'Package', width: 140 }, { field: 'slotStart', headerName: 'Start', width: 220 }, { field: 'status', headerName: 'Status', width: 120 } ] as GridColDef[]}
            height={280}
          />
        )}
      </Paper>
      <DetailsDialog open={!!detail} content={detail || {}} title="Hold Details" onClose={() => setDetail(null)} />
    </Stack>
  );
}
