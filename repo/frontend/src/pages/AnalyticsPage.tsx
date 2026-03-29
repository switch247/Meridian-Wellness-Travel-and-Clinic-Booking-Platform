import { Alert, Button, Grid2 as Grid, Paper, Stack, TextField } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';
import { KpiCard } from '../components/dashboard/KpiCard';

function today() {
  return new Date().toISOString().slice(0, 10);
}

export function AnalyticsPage() {
  const { token } = useAuth();
  const [from, setFrom] = useState(today());
  const [to, setTo] = useState(today());
  const [kpis, setKpis] = useState<Record<string, unknown>>({});
  const [msg, setMsg] = useState<string | null>(null);
  const [reportType, setReportType] = useState('kpi_daily');
  const [scheduledFor, setScheduledFor] = useState(new Date(Date.now() + 3600000).toISOString().slice(0, 16));

  async function load() {
    if (!token) return;
    const out = await api.analyticsKpis(token, { from, to });
    setKpis(out.kpis || {});
  }

  useEffect(() => { load().catch(() => {}); }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Analytics" subtitle="KPI filters, exports, and scheduled local reports." />
      <Paper sx={{ p: 2.5 }}>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
          <TextField type="date" label="From" InputLabelProps={{ shrink: true }} value={from} onChange={(e) => setFrom(e.target.value)} />
          <TextField type="date" label="To" InputLabelProps={{ shrink: true }} value={to} onChange={(e) => setTo(e.target.value)} />
          <Button variant="contained" onClick={() => load()}>Apply Filters</Button>
          <Button variant="outlined" onClick={async () => {
            if (!token) return;
            const out = await api.exportAnalytics(token, { from, to });
            setMsg(`Exported: ${out.path}`);
          }}>Export CSV</Button>
        </Stack>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} sx={{ mt: 1.5 }}>
          <TextField label="Report Type" value={reportType} onChange={(e) => setReportType(e.target.value)} />
          <TextField type="datetime-local" label="Schedule For" InputLabelProps={{ shrink: true }} value={scheduledFor} onChange={(e) => setScheduledFor(e.target.value)} />
          <Button variant="contained" onClick={async () => {
            if (!token) return;
            const out = await api.scheduleReport(token, {
              reportType,
              parameters: { from, to },
              scheduledFor: new Date(scheduledFor).toISOString()
            });
            setMsg(`Scheduled report job #${String((out as { id?: number }).id ?? '')}`);
          }}>Schedule Report</Button>
        </Stack>
      </Paper>
      {msg && <Alert severity="success">{msg}</Alert>}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}><KpiCard label="Booking Volume" value={String(kpis.bookingVolume ?? 0)} /></Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}><KpiCard label="Attendance Rate" value={`${Number(kpis.attendanceRate ?? 0).toFixed(1)}%`} /></Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}><KpiCard label="Repurchase Rate" value={`${Number(kpis.repurchaseRate ?? 0).toFixed(1)}%`} /></Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}><KpiCard label="Refund Rate" value={`${Number(kpis.refundRate ?? 0).toFixed(1)}%`} /></Grid>
      </Grid>
    </Stack>
  );
}
