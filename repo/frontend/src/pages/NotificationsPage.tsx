import { Alert, Button, Chip, CircularProgress, Paper, Stack, Typography } from '@mui/material';
import { useCallback, useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

export function NotificationsPage() {
  const { token } = useAuth();
  const [items, setItems] = useState<Array<Record<string, unknown>>>([]);
  const [markLoading, setMarkLoading] = useState<Record<number, boolean>>({});

  const load = async () => {
    if (!token) return;
    const out = await api.notifications(token);
    setItems(out.items || []);
  };

  useEffect(() => {
    load().catch(() => {});
  }, [token]);

  if (!token) {
    return (
      <Stack spacing={2.5}>
        <SectionHeader title="Notifications" subtitle="Replies, moderation status, and workflow updates." />
        <Alert severity="warning">Please log in to view notifications.</Alert>
      </Stack>
    );
  }

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Notifications" subtitle="Replies, moderation status, and workflow updates." />
      <Paper sx={{ p: 2.5 }}>
        <Stack direction="row" spacing={1}>
          <Button variant="outlined" onClick={() => load().catch(() => {})}>Refresh List</Button>
        </Stack>
        {items.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>
            No notifications yet.
          </Alert>
        ) : (
          <Stack spacing={1.5} sx={{ mt: 1.5 }}>
            {items.map((notification, idx) => {
              const read = Boolean(notification.readAt);
              const id = Number(notification.id ?? 0);
              const isLoading = markLoading[id];
              return (
                <Paper
                  key={idx}
                  variant="outlined"
                  sx={{
                    p: 2,
                    bgcolor: read ? 'grey.100' : 'background.paper',
                    boxShadow: read ? 'none' : '0 0 16px rgba(13, 110, 110, 0.08)'
                  }}
                >
                  <Stack direction="row" justifyContent="space-between" alignItems="flex-start" spacing={2}>
                    <Stack spacing={0.5}>
                      <Typography variant="subtitle1">{String(notification.title)}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {String(notification.body)}
                      </Typography>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Chip label={String(notification.category || 'general')} size="small" />
                        {read && notification.readAt && (
                          <Typography variant="caption" color="text.secondary">
                            Read {new Date(String(notification.readAt)).toLocaleString()}
                          </Typography>
                        )}
                      </Stack>
                    </Stack>
                    <Button
                      size="small"
                      variant={read ? 'outlined' : 'contained'}
                      disabled={read || !id || isLoading}
                      onClick={async () => {
                        if (!token || read || !id) return;
                        setMarkLoading((prev) => ({ ...prev, [id]: true }));
                        try {
                          await api.markNotificationRead(token, id);
                          await load();
                        } finally {
                          setMarkLoading((prev) => ({ ...prev, [id]: false }));
                        }
                      }}
                      startIcon={isLoading ? <CircularProgress size={14} /> : undefined}
                    >
                      {read ? 'Read' : 'Mark as read'}
                    </Button>
                  </Stack>
                </Paper>
              );
            })}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
