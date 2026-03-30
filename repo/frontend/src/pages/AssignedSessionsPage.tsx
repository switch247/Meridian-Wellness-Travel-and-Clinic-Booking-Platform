import { Button, Chip, Grid, Paper, Stack, SxProps, TextField, Typography } from '@mui/material';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

type SessionItem = {
  id: number;
  bookingId?: number;
  travelerId?: number;
  packageId?: number;
  roomId?: number;
  slotStart: string;
  durationMinutes: number;
  status: string;
  sessionNotes?: string;
  sessionNotesSummary?: string;
};

const statusDefinitions: Record<string, { label: string; color: 'default' | 'primary' | 'success' | 'warning' | 'info' }> = {
  scheduled: { label: 'Scheduled', color: 'warning' },
  confirmed: { label: 'Confirmed', color: 'info' },
  checked_in: { label: 'Checked-in', color: 'primary' },
  in_progress: { label: 'In Progress', color: 'info' },
  completed: { label: 'Completed', color: 'success' },
  cancelled: { label: 'Cancelled', color: 'default' }
};
const statusTransitions: Record<string, string[]> = {
  scheduled: ['checked_in'],
  confirmed: ['checked_in'],
  checked_in: ['in_progress'],
  in_progress: ['completed'],
  completed: [],
  cancelled: []
};

const cardStyles: SxProps = {
  border: '1px solid rgba(13,110,110,0.15)',
  borderRadius: 2,
  backgroundColor: 'rgba(255,255,255,0.85)'
};

export function AssignedSessionsPage() {
  const { token, me } = useAuth();
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notesDrafts, setNotesDrafts] = useState<Record<number, string>>({});
  const [actionLoading, setActionLoading] = useState<Record<number, boolean>>({});

  const loadSessions = useCallback(async () => {
    if (!token || !me?.id) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const payload = await api.hostAgenda(token, me.id);
      const normalized = payload.items.map((item) => ({
        id: Number(item.id ?? 0),
        bookingId: item.bookingId ? Number(item.bookingId) : undefined,
        travelerId: item.travelerId ? Number(item.travelerId) : undefined,
        packageId: item.packageId ? Number(item.packageId) : undefined,
        roomId: item.roomId ? Number(item.roomId) : undefined,
        slotStart: String(item.slotStart || ''),
        durationMinutes: Number(item.durationMinutes ?? 0),
        status: String(item.status || 'scheduled'),
        sessionNotes: item.sessionNotes ? String(item.sessionNotes) : undefined,
        sessionNotesSummary: item.sessionNotesSummary ? String(item.sessionNotesSummary) : undefined
      }));
      setSessions(normalized);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load sessions');
    } finally {
      setLoading(false);
    }
  }, [token, me?.id]);

  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  const handleStatusUpdate = useCallback(
    async (bookingId: number, nextStatus: string) => {
      if (!token) return;
      setActionLoading((prev) => ({ ...prev, [bookingId]: true }));
      try {
        const payload: { status: string; notes?: string } = { status: nextStatus };
        const draft = notesDrafts[bookingId];
        if (draft?.trim()) {
          payload.notes = draft.trim();
        }
        await api.updateBookingStatus(token, bookingId, payload);
        await loadSessions();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'status update failed');
      } finally {
        setActionLoading((prev) => ({ ...prev, [bookingId]: false }));
      }
    },
    [token, notesDrafts, loadSessions]
  );

  const refreshSessions = useMemo(() => loadSessions, [loadSessions]);

  return (
    <Stack spacing={3}>
      <SectionHeader
        title="Assigned Sessions"
        subtitle="Coach & clinician view of tomorrow's/next sessions with check-in controls, status transitions, and encrypted notes."
        actions={[
          {
            label: 'Refresh',
            handler: refreshSessions
          }
        ]}
      />

      {error && (
        <Typography color="error.main" sx={{ fontWeight: 500 }}>
          {error}
        </Typography>
      )}

      {loading ? (
        <Stack spacing={2}>
          {[...Array(2)].map((_, idx) => (
            <Paper key={idx} sx={cardStyles}>
              <Stack spacing={1} p={2}>
                <Typography variant="subtitle1" sx={{ width: '65%', height: 16 }}>
                  <span style={{ opacity: 0 }}>Loading</span>
                </Typography>
                <Typography variant="body2" sx={{ width: '40%', height: 16 }}>
                  <span style={{ opacity: 0 }}>Leader</span>
                </Typography>
              </Stack>
            </Paper>
          ))}
        </Stack>
      ) : sessions.length === 0 ? (
        <Paper sx={{ ...cardStyles, px: 3, py: 4 }}>
          <Typography variant="h6" gutterBottom>
            No sessions on record
          </Typography>
          <Typography variant="body2">
            Billing-friendly workflows will automatically surface once a traveler confirms a booking and a
            provider is assigned. Use the catalog to start adding available packages.
          </Typography>
        </Paper>
      ) : (
        <Stack spacing={2}>
          {sessions.map((session) => {
            const statusMeta =
              statusDefinitions[session.status] || { label: session.status, color: 'default' };
            const nextStatuses = statusTransitions[session.status] ?? [];
            const slotStart = session.slotStart ? new Date(session.slotStart) : new Date();
            const slotEnd = new Date(slotStart.getTime() + (session.durationMinutes || 0) * 60000);
            const notesValue = notesDrafts[session.bookingId ?? session.id] ?? session.sessionNotes ?? '';
            return (
              <Paper key={session.id} sx={cardStyles}>
                <Stack spacing={1} p={2}>
                  <Grid container spacing={1} alignItems="center">
                    <Grid item xs>
                      <Typography variant="h6">
                        Session #{session.bookingId ?? session.id} &middot; Traveler {session.travelerId ?? 'TBD'}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Chip label={statusMeta.label} color={statusMeta.color} size="small" />
                    </Grid>
                  </Grid>
                  <Typography variant="body2" color="text.secondary">
                    Package {session.packageId ?? '—'} &middot; Room {session.roomId ?? '—'}
                  </Typography>
                  <Typography variant="body2">
                    {slotStart.toLocaleString()} – {slotEnd.toLocaleTimeString()} ({session.durationMinutes ?? 0} min)
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Notes:&nbsp;
                    {session.sessionNotesSummary ? (
                      session.sessionNotesSummary
                    ) : (
                      <em>No notes yet</em>
                    )}
                  </Typography>
                  <TextField
                    size="small"
                    label="Session notes (encrypted)"
                    placeholder="Add a quick clinical observation or in-progress detail"
                    value={notesValue}
                    onChange={(event) =>
                      setNotesDrafts((prev) => ({
                        ...prev,
                        [session.bookingId ?? session.id]: event.target.value
                      }))
                    }
                    multiline
                    minRows={2}
                  />
                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    {nextStatuses.map((next) => (
                      <Button
                        key={next}
                        variant="contained"
                        size="small"
                        disabled={!session.bookingId || actionLoading[session.bookingId]}
                        onClick={() => session.bookingId && handleStatusUpdate(session.bookingId, next)}
                      >
                        Set {statusDefinitions[next]?.label ?? next}
                      </Button>
                    ))}
                    <Button size="small" variant="outlined" disabled={!session.bookingId}>
                      View traveler session
                    </Button>
                  </Stack>
                </Stack>
              </Paper>
            );
          })}
        </Stack>
      )}
    </Stack>
  );
}
