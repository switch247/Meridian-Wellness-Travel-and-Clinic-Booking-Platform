import { Alert, Box, CircularProgress, Container, Grid2 as Grid, Paper, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { LoginCard } from '../components/LoginCard';

export function HomePage() {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [me, setMe] = useState<{ id: number; username: string; roles: string[]; phone: string; address: string } | null>(null);
  const [catalog, setCatalog] = useState<Array<Record<string, unknown>>>([]);

  async function loadCatalog() {
    const data = await api.catalog();
    setCatalog(data.items);
  }

  useEffect(() => {
    loadCatalog().catch((e: Error) => setError(e.message));
  }, []);

  useEffect(() => {
    if (!token) return;
    api.me(token)
      .then(setMe)
      .catch((e: Error) => {
        setError(e.message);
        if ((e as Error).message.includes('token')) setToken(null);
      });
  }, [token]);

  async function onLogin(value: { username: string; password: string }) {
    setLoading(true);
    setError(null);
    try {
      const result = await api.login(value);
      setToken(result.token);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Container sx={{ py: 6 }}>
      <Typography variant="h4" sx={{ mb: 1 }}>Meridian Wellness Travel Platform</Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
        Offline-ready clinic booking and wellness itinerary operations.
      </Typography>
      {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 5 }}>
          {!token ? (
            <LoginCard loading={loading} onSubmit={onLogin} />
          ) : !me ? (
            <Paper sx={{ p: 3 }}><CircularProgress size={24} /></Paper>
          ) : (
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6">Authenticated Profile</Typography>
              <Typography>Username: {me.username}</Typography>
              <Typography>Roles: {me.roles.join(', ')}</Typography>
              <Typography>Phone: {me.phone}</Typography>
              <Typography>Address: {me.address}</Typography>
            </Paper>
          )}
        </Grid>
        <Grid size={{ xs: 12, md: 7 }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>Published Packages</Typography>
            <Box sx={{ display: 'grid', gap: 1.5 }}>
              {catalog.map((item, idx) => (
                <Paper key={idx} variant="outlined" sx={{ p: 1.5 }}>
                  <Typography variant="subtitle1">{String(item.name ?? '-')}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Destination: {String(item.destination ?? '-')}
                  </Typography>
                  <Typography variant="body2">
                    Inventory: {String(item.inventoryRemaining ?? '-')}
                  </Typography>
                  <Typography variant="body2" color="warning.main">
                    {String(item.blackoutNote ?? '')}
                  </Typography>
                </Paper>
              ))}
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
}
