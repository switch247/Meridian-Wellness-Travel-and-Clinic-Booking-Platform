import { Alert, Button, Grid2 as Grid, Paper, Stack, TextField, Typography } from '@mui/material';
import { FormEvent, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export function LoginPage() {
  const { login, register, loading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState({
    username: '',
    password: '',
    phone: '+15550001111',
    address: '123 Main Street, New York'
  });

  async function onLogin(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await login(form.username, form.password);
      const next = (location.state as { from?: string } | null)?.from || '/';
      navigate(next);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function onRegister() {
    setError(null);
    try {
      await register(form);
      await login(form.username, form.password);
      navigate('/');
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <Grid container sx={{ minHeight: '100vh' }}>
      <Grid size={{ xs: 12, md: 6 }} sx={{ p: { xs: 3, md: 7 }, bgcolor: '#0f766e', color: 'white' }}>
        <Stack spacing={2.5}>
          <Typography variant="overline">Meridian Wellness Platform</Typography>
          <Typography variant="h3">Clinic Scheduling Meets Travel Operations</Typography>
          <Typography sx={{ opacity: 0.88 }}>
            Built for kiosk and office use on a local network with secure role-based workflows and resilient booking holds.
          </Typography>
        </Stack>
      </Grid>
      <Grid size={{ xs: 12, md: 6 }} sx={{ p: { xs: 3, md: 8 }, display: 'grid', placeItems: 'center' }}>
        <Paper sx={{ p: 3.5, width: '100%', maxWidth: 520 }}>
          <Typography variant="h5" sx={{ mb: 2 }}>Sign in</Typography>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <Stack component="form" spacing={2} onSubmit={onLogin}>
            <TextField label="Username" required value={form.username} onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))} />
            <TextField label="Password" type="password" required value={form.password} onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))} />
            <Button type="submit" disabled={loading} variant="contained">{loading ? 'Signing in...' : 'Sign In'}</Button>
            <Button onClick={onRegister} disabled={loading} variant="outlined">Quick Register + Sign In</Button>
          </Stack>
        </Paper>
      </Grid>
    </Grid>
  );
}
