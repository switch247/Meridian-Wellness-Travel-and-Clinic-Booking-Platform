import { Alert, Button, Chip, MenuItem, Paper, Stack, TextField, Typography } from '@mui/material';
import { useMemo, useState } from 'react';

export type BookingPayload = {
  packageId: number;
  hostId: number;
  roomId: number;
  slotStart: string;
  duration: number;
};

export function BookingHoldForm({
  onSubmit,
  packages,
  fetchSlots
}: {
  onSubmit: (payload: BookingPayload) => Promise<void>;
  packages: Array<{ id: number; name: string }>;
  fetchSlots: (input: { hostId: number; roomId: number; day: string; duration: number }) => Promise<Array<{ slotStart: string }>>;
}) {
  const [payload, setPayload] = useState<BookingPayload>({
    packageId: packages[0]?.id || 1,
    hostId: 1,
    roomId: 1,
    slotStart: new Date(Date.now() + 3600000).toISOString().slice(0, 16),
    duration: 45
  });
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [slots, setSlots] = useState<Array<{ slotStart: string }>>([]);

  const minTime = useMemo(() => new Date().toISOString().slice(0, 16), []);

  return (
    <Paper sx={{ p: 2.5 }}>
      <Stack spacing={2}>
        {status && <Alert severity="success">{status}</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
        <TextField
          select
          label="Package"
          value={payload.packageId}
          onChange={(e) => setPayload((p) => ({ ...p, packageId: Number(e.target.value) }))}
        >
          {packages.map((pkg) => (
            <MenuItem key={pkg.id} value={pkg.id}>{pkg.name}</MenuItem>
          ))}
        </TextField>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
          <TextField label="Host ID" type="number" value={payload.hostId} onChange={(e) => setPayload((p) => ({ ...p, hostId: Number(e.target.value) }))} fullWidth />
          <TextField label="Room ID" type="number" value={payload.roomId} onChange={(e) => setPayload((p) => ({ ...p, roomId: Number(e.target.value) }))} fullWidth />
        </Stack>
        <TextField
          label="Slot Start"
          type="datetime-local"
          value={payload.slotStart}
          onChange={(e) => setPayload((p) => ({ ...p, slotStart: e.target.value }))}
          inputProps={{ min: minTime }}
        />
        <Button
          variant="outlined"
          onClick={async () => {
            const day = payload.slotStart.slice(0, 10);
            const available = await fetchSlots({ hostId: payload.hostId, roomId: payload.roomId, day, duration: payload.duration });
            setSlots(available);
            if (available[0]?.slotStart) {
              setPayload((p) => ({ ...p, slotStart: new Date(available[0].slotStart).toISOString().slice(0, 16) }));
            }
          }}
        >
          Load Available Slots
        </Button>
        {slots.length > 0 ? (
          <Stack spacing={1}>
            <Typography variant="body2" color="text.secondary">Next available windows</Typography>
            <Stack direction="row" spacing={1} flexWrap="wrap">
              {slots.map((s) => {
                const slotDate = new Date(s.slotStart);
                return (
                  <Chip
                    key={s.slotStart}
                    label={slotDate.toLocaleString()}
                    color={slotDate.toISOString().slice(0, 10) === payload.slotStart.slice(0, 10) ? 'primary' : 'default'}
                    clickable
                    onClick={() => setPayload((p) => ({
                      ...p,
                      slotStart: slotDate.toISOString().slice(0, 16)
                    }))}
                  />
                );
              })}
            </Stack>
          </Stack>
        ) : (
          <Typography variant="caption" color="text.secondary">Use \"Load Available Slots\" to see capacity.</Typography>
        )}
        <TextField
          select
          label="Duration"
          value={payload.duration}
          onChange={(e) => setPayload((p) => ({ ...p, duration: Number(e.target.value) }))}
        >
          {[30, 45, 60].map((d) => <MenuItem key={d} value={d}>{d} minutes</MenuItem>)}
        </TextField>
        <Button
          variant="contained"
          disabled={loading}
          onClick={async () => {
            setLoading(true);
            setError(null);
            setStatus(null);
            try {
              await onSubmit({ ...payload, slotStart: new Date(payload.slotStart).toISOString() });
              setStatus('Reservation hold created successfully.');
            } catch (err) {
              setError((err as Error).message);
            } finally {
              setLoading(false);
            }
          }}
        >
          {loading ? 'Placing hold...' : 'Place Reservation Hold'}
        </Button>
      </Stack>
    </Paper>
  );
}
