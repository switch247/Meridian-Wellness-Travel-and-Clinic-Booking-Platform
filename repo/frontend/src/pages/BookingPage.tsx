import {
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid2 as Grid,
  Paper,
  Stack,
  Typography
} from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { BookingHoldForm } from '../components/booking/BookingHoldForm';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';

export function BookingPage() {
  const { token } = useAuth();
  const [packages, setPackages] = useState<Array<{ id: number; name: string }>>([]);
  const [open, setOpen] = useState(false);
  const [holds, setHolds] = useState<Array<Record<string, unknown>>>([]);
  const [history, setHistory] = useState<Array<Record<string, unknown>>>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api.catalog().then((r) => {
      const pkgs = r.items.map((it, idx) => ({ id: Number(it.id ?? idx + 1), name: String(it.name ?? 'Package') }));
      const deduped = Array.from(new Map(pkgs.map((p) => [p.id, p])).values());
      setPackages(deduped.length ? deduped : [{ id: 1, name: 'Fallback Package' }]);
    });
    if (token) {
      api.listHolds(token).then((r) => setHolds(r.items || [])).catch(() => {});
      api.listHistory(token).then((r) => setHistory(r.items || [])).catch(() => {});
    }
  }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Booking Studio" subtitle="Create temporary reservation holds with live conflict checks." />
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <Button variant="contained" onClick={() => setOpen(true)}>Create Reservation Hold</Button>
          <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>Create Reservation Hold</DialogTitle>
          <DialogContent>
            <BookingHoldForm
              packages={packages}
              fetchSlots={async (input) => {
                if (!token) return [];
                const out = await api.availableSlots(token, input);
                return (out.items || []).map((i) => ({ slotStart: String(i.slotStart) }));
              }}
              onSubmit={async (payload) => {
                if (!token) throw new Error('Please login first');
                await api.placeHold(token, payload);
                setOpen(false);
                setLoading(true);
                await api.listHolds(token).then((r) => setHolds(r.items || [])).catch(() => {});
                await api.listHistory(token).then((r) => setHistory(r.items || [])).catch(() => {});
                setLoading(false);
              }}
            />
          </DialogContent>
        </Dialog>
      </Grid>
        <Grid size={{ xs: 12, lg: 5 }}>
          <Paper sx={{ p: 2.5, height: '100%' }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Rules Applied</Typography>
            <Typography variant="body2" color="text.secondary">
              - Multi-resource conflict validation (host + room + chair) enforced on every hold.
            </Typography>
            <Typography variant="body2" color="text.secondary">
              - Limited inventory decrements once the hold is confirmed, preventing oversell.
            </Typography>
            <Typography variant="body2" color="text.secondary">
              - 10-minute TTL automatically releases if confirmation is delayed.
            </Typography>
            <Typography variant="body2" color="text.secondary">
              - Optimistic versioning guards against parallel confirmations.
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Typography variant="h6">Active Holds</Typography>
          <Stack spacing={1.1} sx={{ mt: 1 }}>
            {loading && <CircularProgress size={20} />}
            {holds.length === 0 ? (
              <Alert severity="info">No active holds yet. Use \"Create Reservation Hold\" to reserve a slot.</Alert>
            ) : (
              holds.map((h, idx) => (
                <Box key={`hold-${idx}`} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 1.2 }}>
                  <Typography variant="body2">
                    #{String(h.id)} · Package {String(h.packageId)} · {String(h.status)} · Slot {new Date(String(h.slotStart || '')).toLocaleString()}
                  </Typography>
                </Box>
              ))
            )}
          </Stack>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Typography variant="h6">Recent Confirmed</Typography>
          <Stack spacing={1.1} sx={{ mt: 1 }}>
            {history.length === 0 ? (
              <Alert severity="info">No confirmed bookings yet.</Alert>
            ) : (
              history.slice(0, 3).map((h, idx) => (
                <Box key={`hist-${idx}`} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 1.2 }}>
                  <Typography variant="body2">
                    Booking #{String(h.id)} · {String(h.status)} · {new Date(String(h.slotStart || '')).toLocaleString()}
                  </Typography>
                </Box>
              ))
            )}
          </Stack>
        </Grid>
      </Grid>
    </Stack>
  );
}
