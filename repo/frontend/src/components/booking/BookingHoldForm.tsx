import { Alert, Button, Chip, MenuItem, Paper, Stack, TextField, Typography } from '@mui/material';
import { useEffect, useMemo, useState } from 'react';

export type BookingPayload = {
  packageId: number;
  hostId: number;
  roomId: number;
  chairId?: number;
  slotStart: string;
  duration: number;
};

export function BookingHoldForm({
  onSubmit,
  packages,
  fetchSlots,
  fetchChairs,
  hosts,
  rooms
}: {
  onSubmit: (payload: BookingPayload) => Promise<void>;
  packages: Array<{ id: number; name: string }>;
  fetchSlots: (input: { hostId: number; roomId: number; chairId?: number; day: string; duration: number }) => Promise<Array<{ slotStart: string }>>;
  fetchChairs: (roomId: number) => Promise<Array<{ id: number; name: string }>>;
  hosts: Array<{ id: number; username: string }>;
  rooms: Array<{ id: number; name: string; chairsCount?: number }>;
}) {
  const [payload, setPayload] = useState<BookingPayload>({
    packageId: packages[0]?.id || 0,
    hostId: hosts[0]?.id || 0,
    roomId: rooms[0]?.id || 0,
    slotStart: new Date(Date.now() + 3600000).toISOString().slice(0, 16), // 1 hour from now 
    duration: 45
  });
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [slots, setSlots] = useState<Array<{ slotStart: string }>>([]);
  const [slotsLoading, setSlotsLoading] = useState(false);
  const [chairs, setChairs] = useState<Array<{ id: number; name: string }>>([]);
  const [chairsLoading, setChairsLoading] = useState(false);

  const minTime = useMemo(() => new Date(Date.now() + 60 * 60 * 1000).toISOString().slice(0, 16), []);

  const requiresChair = chairs.length > 0;
  const canLoadSlots = payload.hostId > 0 && payload.roomId > 0 && (!requiresChair || Number(payload.chairId) > 0);
  const canSubmit = payload.packageId > 0 && payload.hostId > 0 && payload.roomId > 0 && (!requiresChair || Number(payload.chairId) > 0);

  useEffect(() => {
    setPayload((prev) => ({
      ...prev,
      packageId: prev.packageId > 0 ? prev.packageId : (packages[0]?.id || 0),
      hostId: prev.hostId > 0 ? prev.hostId : (hosts[0]?.id || 0),
      roomId: prev.roomId > 0 ? prev.roomId : (rooms[0]?.id || 0)
    }));
  }, [packages, hosts, rooms]);

  useEffect(() => {
    if (payload.roomId <= 0) {
      setChairs([]);
      setPayload((prev) => ({ ...prev, chairId: undefined }));
      return;
    }
    setChairsLoading(true);
    fetchChairs(payload.roomId)
      .then((items) => {
        setChairs(items);
        setPayload((prev) => ({
          ...prev,
          chairId: items.length > 0 ? (prev.chairId && items.some((c) => c.id === prev.chairId) ? prev.chairId : items[0].id) : undefined
        }));
      })
      .catch(() => {
        setChairs([]);
        setPayload((prev) => ({ ...prev, chairId: undefined }));
      })
      .finally(() => setChairsLoading(false));
  }, [payload.roomId]);

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
          <TextField
            select
            label="Host"
            value={payload.hostId}
            onChange={(e) => setPayload((p) => ({ ...p, hostId: Number(e.target.value) }))}
            fullWidth
          >
            {hosts.map((host) => (
              <MenuItem key={host.id} value={host.id}>{host.username}</MenuItem>
            ))}
          </TextField>
          <TextField
            select
            label="Room"
            value={payload.roomId}
            onChange={(e) => setPayload((p) => ({ ...p, roomId: Number(e.target.value) }))}
            fullWidth
          >
            {rooms.map((room) => (
              <MenuItem key={room.id} value={room.id}>
                {room.name}{typeof room.chairsCount === 'number' ? ` (${room.chairsCount} chairs)` : ''}
              </MenuItem>
            ))}
          </TextField>
        </Stack>
        <TextField
          select
          label="Chair"
          value={payload.chairId ?? ''}
          onChange={(e) => setPayload((p) => ({ ...p, chairId: e.target.value ? Number(e.target.value) : undefined }))}
          disabled={chairsLoading || chairs.length === 0}
          helperText={chairs.length === 0 ? 'No explicit chair records for this room.' : 'Select chair for chair-level conflict checks.'}
        >
          {chairs.length === 0 ? (
            <MenuItem value="">No chairs available</MenuItem>
          ) : (
            chairs.map((chair) => (
              <MenuItem key={chair.id} value={chair.id}>{chair.name}</MenuItem>
            ))
          )}
        </TextField>
        <TextField
          label="Slot Start"
          type="datetime-local"
          value={payload.slotStart}
          onChange={(e) => setPayload((p) => ({ ...p, slotStart: e.target.value }))}
          inputProps={{ min: minTime }}
        />
        <Button
          variant="outlined"
          disabled={slotsLoading || !canLoadSlots}
          onClick={async () => {
            setSlotsLoading(true);
            try {
              const day = payload.slotStart.slice(0, 10);
              const available = await fetchSlots({ hostId: payload.hostId, roomId: payload.roomId, chairId: payload.chairId, day, duration: payload.duration });
              setSlots(available);
              if (available[0]?.slotStart) {
                setPayload((p) => ({ ...p, slotStart: new Date(available[0].slotStart).toISOString().slice(0, 16) }));
              }
            } finally {
              setSlotsLoading(false);
            }
          }}
        >
          {slotsLoading ? 'Loading...' : 'Load Available Slots'}
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
        ) : slotsLoading ? (
          <Typography variant="caption" color="text.secondary">Loading slots...</Typography>
        ) : slots.length === 0 && !slotsLoading ? (
          <Typography variant="caption" color="text.secondary">No available slots for the selected date and duration.</Typography>
        ) : (
          <Typography variant="caption" color="text.secondary">Use "Load Available Slots" to see capacity.</Typography>
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
          disabled={loading || !canSubmit}
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
