import { Alert, Button, Card, CardContent, Grid2 as Grid, MenuItem, Paper, Stack, TextField, Typography } from '@mui/material';
import EventAvailableRoundedIcon from '@mui/icons-material/EventAvailableRounded';
import FavoriteRoundedIcon from '@mui/icons-material/FavoriteRounded';
import InsightsRoundedIcon from '@mui/icons-material/InsightsRounded';
import PaidRoundedIcon from '@mui/icons-material/PaidRounded';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';
import { KpiCard } from '../components/dashboard/KpiCard';

function today() {
  return new Date().toISOString().slice(0, 10);
}

export function AnalyticsPage() {
  const { token, me } = useAuth();
  const isAdmin = me?.roles?.includes('admin');
  const [from, setFrom] = useState(today());
  const [to, setTo] = useState(today());
  const [kpis, setKpis] = useState<Record<string, unknown>>({});
  const [msg, setMsg] = useState<string | null>(null);
  const [reportType, setReportType] = useState('kpi_daily');
  const [scheduledFor, setScheduledFor] = useState(new Date(Date.now() + 3600000).toISOString().slice(0, 16));
  const [providerId, setProviderId] = useState<number | undefined>(undefined);
  const [packageId, setPackageId] = useState<number | undefined>(undefined);
  const [providers, setProviders] = useState<Array<{ id: number; username: string }>>([]);
  const [packages, setPackages] = useState<Array<{ id: number; name: string }>>([]);
  const [reportJobs, setReportJobs] = useState<Array<Record<string, unknown>>>([]);

  const loadKpis = async () => {
    if (!token) return;
    const out = await api.analyticsKpis(token, { from, to, providerId, packageId });
    setKpis(out.kpis || {});
  };

  const loadJobs = async () => {
    if (!token) return;
    const out = await api.reportJobs(token);
    setReportJobs(out.items || []);
  };

  const loadMasters = async () => {
    if (!token) return;
    const catalog = await api.catalog();
    setPackages((catalog.items || []).map((pkg) => ({
      id: Number(pkg.id),
      name: String(pkg.name || `Package ${pkg.id}`)
    })));
    const coachUsers = await api.adminUsers(token, 'coach');
    const clinicianUsers = await api.adminUsers(token, 'clinician');
    const uniqueProviders = [...coachUsers.items, ...clinicianUsers.items];
    setProviders(uniqueProviders.map((u) => ({ id: Number(u.id), username: String(u.username || '?') })));
  };

  useEffect(() => {
    loadMasters().catch(() => {});
    loadJobs().catch(() => {});
  }, [token]);

  useEffect(() => {
    loadKpis().catch(() => {});
  }, [token, from, to, providerId, packageId]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Analytics" subtitle="KPI filters, exports, and scheduled local reports." />
      <Paper sx={{ p: 2.5 }}>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
          <TextField type="date" label="From" InputLabelProps={{ shrink: true }} value={from} onChange={(e) => setFrom(e.target.value)} />
          <TextField type="date" label="To" InputLabelProps={{ shrink: true }} value={to} onChange={(e) => setTo(e.target.value)} />
          <TextField
            select
            label="Provider"
            value={providerId ?? ''}
            onChange={(e) => setProviderId(e.target.value ? Number(e.target.value) : undefined)}
            sx={{ minWidth: 200 }}
          >
            <MenuItem value="">Any provider</MenuItem>
            {providers.map((provider) => (
              <MenuItem key={provider.id} value={provider.id}>{provider.username}</MenuItem>
            ))}
          </TextField>
          <TextField
            select
            label="Package"
            value={packageId ?? ''}
            onChange={(e) => setPackageId(e.target.value ? Number(e.target.value) : undefined)}
            sx={{ minWidth: 200 }}
          >
            <MenuItem value="">Any package</MenuItem>
            {packages.map((pkg) => (
              <MenuItem key={pkg.id} value={pkg.id}>{pkg.name}</MenuItem>
            ))}
          </TextField>
          <Button variant="contained" onClick={() => loadKpis().catch(() => {})}>Apply Filters</Button>
          <Button variant="outlined" onClick={async () => {
            if (!token) return;
            const out = await api.exportAnalytics(token, { from, to, providerId, packageId });
            setMsg(`Exported to ${out.path}`);
          }}>Export CSV</Button>
        </Stack>
        {isAdmin && (
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} sx={{ mt: 1.5 }}>
            <TextField label="Report Type" value={reportType} onChange={(e) => setReportType(e.target.value)} sx={{ minWidth: 200 }} />
            <TextField
              type="datetime-local"
              label="Schedule For"
              InputLabelProps={{ shrink: true }}
              value={scheduledFor}
              onChange={(e) => setScheduledFor(e.target.value)}
              sx={{ minWidth: 250 }}
            />
            <Button variant="contained" onClick={async () => {
              if (!token) return;
              const out = await api.scheduleReport(token, {
                reportType,
                parameters: { from, to, providerId, packageId },
                scheduledFor: new Date(scheduledFor).toISOString()
              });
              setMsg(`Scheduled job #${String((out as { id?: number }).id ?? '')}`);
              loadJobs();
            }}>Schedule Report</Button>
          </Stack>
        )}
        <Button variant="text" onClick={() => loadJobs().catch(() => {})} sx={{ mt: 1 }}>Refresh Jobs</Button>
      </Paper>
      {msg && <Alert severity="success">{msg}</Alert>}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard label="Booking Volume" value={String(kpis.bookingVolume ?? 0)} icon={<EventAvailableRoundedIcon color="primary" />} />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard label="Attendance Rate" value={`${Number(kpis.attendanceRate ?? 0).toFixed(1)}%`} icon={<InsightsRoundedIcon color="secondary" />} tone="secondary" />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard label="Repurchase Rate" value={`${Number(kpis.repurchaseRate ?? 0).toFixed(1)}%`} icon={<FavoriteRoundedIcon color="success" />} tone="success" />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <KpiCard label="Refund Rate" value={`${Number(kpis.refundRate ?? 0).toFixed(1)}%`} icon={<PaidRoundedIcon color="primary" />} />
        </Grid>
      </Grid>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Scheduled Reports</Typography>
        {reportJobs.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1 }}>No scheduled reports yet.</Alert>
        ) : (
          <Stack spacing={1.3} sx={{ mt: 1 }}>
            {reportJobs.map((job, idx) => (
              <Card key={idx} variant="outlined">
                <CardContent>
                  <Typography variant="subtitle2">Job #{String(job.id)}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Type: {String(job.reportType)} · Status: {String(job.status)} · Scheduled: {new Date(String((job.scheduledFor ?? job.createdAt) || '')).toLocaleString()}
                  </Typography>
                </CardContent>
              </Card>
            ))}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
