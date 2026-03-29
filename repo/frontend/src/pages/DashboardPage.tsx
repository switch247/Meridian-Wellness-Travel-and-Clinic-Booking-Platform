import { Alert, Box, Grid2 as Grid, Paper, Stack, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { KpiCard } from '../components/dashboard/KpiCard';
import { SectionHeader } from '../components/common/SectionHeader';
import EventAvailableRoundedIcon from '@mui/icons-material/EventAvailableRounded';
import FavoriteRoundedIcon from '@mui/icons-material/FavoriteRounded';
import InsightsRoundedIcon from '@mui/icons-material/InsightsRounded';
import PaidRoundedIcon from '@mui/icons-material/PaidRounded';
import { useAuth } from '../context/AuthContext';

export function DashboardPage() {
  const { token, me } = useAuth();
  const [holds, setHolds] = useState<Array<Record<string, unknown>>>([]);
  const [history, setHistory] = useState<Array<Record<string, unknown>>>([]);
  const [kpis, setKpis] = useState<Record<string, unknown>>({});

  useEffect(() => {
    if (!token) return;
    api.listHolds(token).then((r) => setHolds(r.items || [])).catch(() => {});
    api.listHistory(token).then((r) => setHistory(r.items || [])).catch(() => {});
    if (me?.roles?.includes('operations') || me?.roles?.includes('admin')) {
      const today = new Date().toISOString().slice(0, 10);
      api.analyticsKpis(token, { from: today, to: today }).then((r) => setKpis(r.kpis || {})).catch(() => {});
    }
  }, [token]);

  return (
    <Stack spacing={3}>
      <SectionHeader
        title="Operational Dashboard"
        subtitle="Live local-network KPIs for bookings, attendance, and revenue quality."
      />
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard icon={<EventAvailableRoundedIcon color="primary" />} label="Booking Volume" value={String(kpis.bookingVolume ?? holds.length)} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard icon={<InsightsRoundedIcon color="secondary" />} label="Attendance Rate" value={`${Number(kpis.attendanceRate ?? 0).toFixed(1)}%`} tone="secondary" />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard icon={<FavoriteRoundedIcon color="success" />} label="Repurchase Rate" value={`${Number(kpis.repurchaseRate ?? 0).toFixed(1)}%`} tone="success" />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard icon={<PaidRoundedIcon color="primary" />} label="Refund Rate" value={`${Number(kpis.refundRate ?? 0).toFixed(1)}%`} />
        </Grid>
      </Grid>

      {me?.roles?.includes('traveler') && (
        <Paper sx={{ p: 2.5 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>My Reservations</Typography>
          {holds.length === 0 && history.length === 0 ? (
            <Alert severity="info">No reservations yet. Go to Catalog or Booking to start.</Alert>
          ) : (
            <Box sx={{ display: 'grid', gap: 1.2 }}>
              {holds.slice(0, 3).map((h, i) => (
                <Typography key={i}>Hold #{String(h.id)} | package {String(h.packageId)} | status {String(h.status)}</Typography>
              ))}
              {history.slice(0, 3).map((h, i) => (
                <Typography key={`hist-${i}`}>Booking #{String(h.id)} | package {String(h.packageId)} | status {String(h.status)}</Typography>
              ))}
            </Box>
          )}
        </Paper>
      )}
    </Stack>
  );
}
