import {
  Alert,
  Box,
  Card,
  CardContent,
  CardMedia,
  Chip,
  Grid2 as Grid,
  Paper,
  Stack,
  Typography
} from '@mui/material';
import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { KpiCard } from '../components/dashboard/KpiCard';
import { SectionHeader } from '../components/common/SectionHeader';
import EventAvailableRoundedIcon from '@mui/icons-material/EventAvailableRounded';
import FavoriteRoundedIcon from '@mui/icons-material/FavoriteRounded';
import InsightsRoundedIcon from '@mui/icons-material/InsightsRounded';
import PaidRoundedIcon from '@mui/icons-material/PaidRounded';
import LandscapeRoundedIcon from '@mui/icons-material/LandscapeRounded';
import ForumRoundedIcon from '@mui/icons-material/ForumRounded';
import { useAuth } from '../context/AuthContext';

export function DashboardPage() {
  const { token, me } = useAuth();
  const [holds, setHolds] = useState<Array<Record<string, unknown>>>([]);
  const [history, setHistory] = useState<Array<Record<string, unknown>>>([]);
  const [kpis, setKpis] = useState<Record<string, unknown>>({});
  const [routes, setRoutes] = useState<Array<Record<string, unknown>>>([]);
  const [postsCount, setPostsCount] = useState(0);
  const [packages, setPackages] = useState<Array<Record<string, unknown>>>([]);

  useEffect(() => {
    if (!token) return;
    api.listHolds(token).then((r) => setHolds(r.items || [])).catch(() => {});
    api.listHistory(token).then((r) => setHistory(r.items || [])).catch(() => {});
    api.catalog().then((r) => setPackages(r.items || [])).catch(() => {});
    api.routes().then((r) => setRoutes(r.items || [])).catch(() => {});
    api.communityPosts(token).then((r) => setPostsCount((r.items || []).length)).catch(() => {});
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

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 2.5 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Traveler Snapshot</Typography>
            {holds.length + history.length === 0 ? (
              <Alert severity="info">No reservations yet. Head to Catalog or Booking to reserve your first session.</Alert>
            ) : (
              <Stack spacing={1.2}>
                {holds.slice(0, 3).map((h, i) => (
                  <Typography key={`hold-${i}`} variant="body2">
                    Hold #{String(h.id)} · Package #{String(h.packageId)} · Status {String(h.status)}
                  </Typography>
                ))}
                {history.slice(0, 2).map((h, i) => (
                  <Typography key={`hist-${i}`} variant="body2">
                    Booking #{String(h.id)} · Package #{String(h.packageId)} · Confirmed {String(h.status)}
                  </Typography>
                ))}
              </Stack>
            )}
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Paper sx={{ p: 2.5 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Community Pulse</Typography>
            <Stack direction="row" spacing={1}>
              <Chip icon={<ForumRoundedIcon />} label={`${postsCount} active posts`} color="info" />
              <Chip label={`${routes.length} curated routes`} color="secondary" />
            </Stack>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
              Community replies and moderation outcomes feed this dashboard via in-app notifications.
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      <Typography variant="h5" sx={{ mt: 1 }}>Travel Discovery</Typography>
      <Grid container spacing={2}>
        {(packages.slice(0, 3) || []).map((pkg, idx) => (
          <Grid key={idx} size={{ xs: 12, md: 4 }}>
            <Card sx={{ minHeight: 220, display: 'flex', flexDirection: 'column' }}>
              <CardMedia
                component="img"
                height="140"
                image={String(pkg.imagePath || '/placeholder.jpg')}
                alt={String(pkg.name || 'Package')}
              />
              <CardContent sx={{ flexGrow: 1 }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                  {String(pkg.name || `Package ${pkg.id}`)}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                  {String(pkg.description || '7-day curated wellness itinerary.')}
                </Typography>
                <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                  <Chip label={`Inventory ${pkg.inventoryRemaining ?? 'N/A'}`} size="small" />
                  <Chip label={`Status ${pkg.published ? 'Published' : 'Draft'}`} size="small" />
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Stack>
  );
}
