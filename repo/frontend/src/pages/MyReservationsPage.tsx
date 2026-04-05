import { Alert, Button, Paper, Stack, Typography, Dialog, DialogTitle, DialogContent, DialogActions, Menu, MenuItem, IconButton } from '@mui/material';
import MoreVertIcon from '@mui/icons-material/MoreVert';
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
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<any>(null);

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
            rows={(holds as any[]).map(h => ({
              id: Number(h.id),
              packageId: String(h.packageId || 'N/A'),
              status: String(h.status || 'Unknown'),
              slotStart: new Date(String(h.slotStart || '')).toLocaleString(),
              _raw: h
            }))}
            columns={[
              { field: 'id', headerName: 'ID', width: 70 },
              { field: 'packageId', headerName: 'Package', flex: 1, minWidth: 120 },
              { field: 'slotStart', headerName: 'Start Time', flex: 1, minWidth: 180 },
              { field: 'status', headerName: 'Status', width: 100 },
              {
                field: 'actions',
                headerName: '',
                width: 50,
                sortable: false,
                renderCell: (p: any) => (
                  p.row.status === 'active' ? (
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        setMenuAnchor(e.currentTarget);
                        setSelectedRow(p.row);
                      }}
                    >
                      <MoreVertIcon />
                    </IconButton>
                  ) : (
                    <Typography variant="caption" color="text.secondary">
                      {p.row.status}
                    </Typography>
                  )
                )
              }
            ] as GridColDef[]}
            height={300}
            sx={{ width: '100%' }}
          />
        )}
      </Paper>
      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={() => setMenuAnchor(null)}
      >
        <MenuItem onClick={() => {
          setDetail(selectedRow?._raw);
          setMenuAnchor(null);
        }}>
          Details
        </MenuItem>
        <MenuItem onClick={async () => {
          if (!token || !selectedRow) return;
          const version = Number(selectedRow._raw?.version ?? 0);
          if (version <= 0) {
            window.alert('Unable to confirm: version data is missing. Refresh and try again.');
            setMenuAnchor(null);
            return;
          }
          await api.confirmHold(token, { holdId: Number(selectedRow.id), version });
          await load();
          setMenuAnchor(null);
        }}>
          Confirm
        </MenuItem>
        <MenuItem onClick={() => {
          setConfirm({ open: true, id: Number(selectedRow?.id), message: 'Cancel this hold?' });
          setMenuAnchor(null);
        }}>
          Cancel
        </MenuItem>
      </Menu>
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
            rows={(history as any[]).map(h => ({
              id: Number(h.id),
              packageId: String(h.packageId || 'N/A'),
              status: String(h.status || 'Unknown'),
              slotStart: new Date(String(h.slotStart || '')).toLocaleString()
            }))}
            columns={[
              { field: 'id', headerName: 'ID', width: 70 },
              { field: 'packageId', headerName: 'Package', flex: 1, minWidth: 150 },
              { field: 'slotStart', headerName: 'Start Time', flex: 1, minWidth: 200 },
              { field: 'status', headerName: 'Status', width: 120 }
            ] as GridColDef[]}
            height={300}
            sx={{ width: '100%' }}
          />
        )}
      </Paper>
      <DetailsDialog open={!!detail} content={detail || {}} title="Hold Details" onClose={() => setDetail(null)} />
    </Stack>
  );
}
