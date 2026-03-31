import { Alert, Button, FormControl, InputLabel, MenuItem, Paper, Select, Stack, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import EntityTable from '../components/common/EntityTable';
import { GridColDef } from '@mui/x-data-grid';

export function OpsSchedulingPage() {
  const { token } = useAuth();
  const [hostId, setHostId] = useState('');
  const [roomId, setRoomId] = useState('');
  const [hosts, setHosts] = useState<Array<Record<string, unknown>>>([]);
  const [rooms, setRooms] = useState<Array<Record<string, unknown>>>([]);
  const [hostAgenda, setHostAgenda] = useState<Array<Record<string, unknown>>>([]);
  const [roomAgenda, setRoomAgenda] = useState<Array<Record<string, unknown>>>([]);

  useEffect(() => {
    if (token) {
      api.listHosts(token).then(r => setHosts(r.items || [])).catch(() => {});
      api.listRooms(token).then(r => setRooms(r.items || [])).catch(() => {});
    }
  }, [token]);

  async function load() {
    if (!token || !hostId || !roomId) return;
    const h = await api.hostAgenda(token, Number(hostId));
    const r = await api.roomAgenda(token, Number(roomId));
    setHostAgenda(h.items || []);
    setRoomAgenda(r.items || []);
  }

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Scheduling Ops" subtitle="Capacity and agenda visibility for operations/admin." />
      <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
        <FormControl fullWidth>
          <InputLabel>Host</InputLabel>
          <Select value={hostId} label="Host" onChange={(e) => setHostId(e.target.value)}>
            {hosts.map(h => <MenuItem key={h.id} value={String(h.id)}>{String(h.name || h.username || `Host ${h.id}`)}</MenuItem>)}
          </Select>
        </FormControl>
        <FormControl fullWidth>
          <InputLabel>Room</InputLabel>
          <Select value={roomId} label="Room" onChange={(e) => setRoomId(e.target.value)}>
            {rooms.map((r) => (
              <MenuItem key={String(r.id)} value={String(r.id)}>
                {String(r.name || `Room ${r.id}`)}
                {typeof r.chairsCount === 'number' ? ` (${String(r.chairsCount)} chairs)` : ''}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Button variant="contained" onClick={() => load()} disabled={!hostId || !roomId}>Load Agendas</Button>
      </Stack>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Host Agenda</Typography>
        {hostAgenda.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1 }}>No host sessions found.</Alert>
        ) : (
          <EntityTable rows={(hostAgenda as any[]).map(it => ({ id: Number(it.id), travelerId: it.travelerId, slotStart: it.slotStart, status: it.status }))}
            columns={[{ field: 'id', headerName: 'ID', width: 90 }, { field: 'travelerId', headerName: 'Traveler', width: 160 }, { field: 'slotStart', headerName: 'Start', width: 220 }, { field: 'status', headerName: 'Status', width: 120 }] as GridColDef[]} height={300} />
        )}
      </Paper>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Room Agenda</Typography>
        {roomAgenda.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1 }}>No room sessions found.</Alert>
        ) : (
          <EntityTable rows={(roomAgenda as any[]).map(it => ({ id: Number(it.id), hostId: it.hostId, slotStart: it.slotStart, status: it.status }))}
            columns={[{ field: 'id', headerName: 'ID', width: 90 }, { field: 'hostId', headerName: 'Host', width: 160 }, { field: 'slotStart', headerName: 'Start', width: 220 }, { field: 'status', headerName: 'Status', width: 120 }] as GridColDef[]} height={300} />
        )}
      </Paper>
    </Stack>
  );
}
