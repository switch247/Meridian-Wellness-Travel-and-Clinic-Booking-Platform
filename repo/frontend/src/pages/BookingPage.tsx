import { Grid2 as Grid, Paper, Stack, Typography, Button, Dialog, DialogTitle, DialogContent } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { BookingHoldForm } from '../components/booking/BookingHoldForm';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';

export function BookingPage() {
  const { token } = useAuth();
  const [packages, setPackages] = useState<Array<{ id: number; name: string }>>([]);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    api.catalog().then((r) => {
      const pkgs = r.items.map((it, idx) => ({ id: Number(it.id ?? idx + 1), name: String(it.name ?? 'Package') }));
      const deduped = Array.from(new Map(pkgs.map((p) => [p.id, p])).values());
      setPackages(deduped.length ? deduped : [{ id: 1, name: 'Fallback Package' }]);
    });
  }, []);

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
                }}
              />
            </DialogContent>
          </Dialog>
        </Grid>
        <Grid size={{ xs: 12, lg: 5 }}>
          <Paper sx={{ p: 2.5, height: '100%' }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Rules Applied</Typography>
            <Typography variant="body2" color="text.secondary">- Multi-resource conflict validation (host + room)</Typography>
            <Typography variant="body2" color="text.secondary">- Quota decrement on successful hold</Typography>
            <Typography variant="body2" color="text.secondary">- 10-minute expiry auto-release on backend</Typography>
            <Typography variant="body2" color="text.secondary">- Optimistic versioning for consistency</Typography>
          </Paper>
        </Grid>
      </Grid>
    </Stack>
  );
}
