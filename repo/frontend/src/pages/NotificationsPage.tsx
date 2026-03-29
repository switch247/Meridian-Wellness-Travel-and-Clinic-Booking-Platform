import { Alert, Button, Paper, Stack, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

export function NotificationsPage() {
  const { token } = useAuth();
  const [items, setItems] = useState<Array<Record<string, unknown>>>([]);

  async function load() {
    if (!token) return;
    const out = await api.notifications(token);
    setItems(out.items || []);
  }

  useEffect(() => { load().catch(() => {}); }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Notifications" subtitle="Replies, moderation status, and workflow updates." />
      <Paper sx={{ p: 2.5 }}>
        <Button variant="outlined" onClick={() => load()}>Refresh</Button>
        {items.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>No notifications yet.</Alert>
        ) : (
          <Stack spacing={1.2} sx={{ mt: 1.5 }}>
            {items.map((n, i) => (
              <Paper key={i} variant="outlined" sx={{ p: 1.5 }}>
                <Typography variant="subtitle2">{String(n.title)}</Typography>
                <Typography variant="body2" color="text.secondary">{String(n.body)}</Typography>
              </Paper>
            ))}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
