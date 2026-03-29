import { Alert, Button, Paper, Stack, TextField, Typography } from '@mui/material';
import { useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import EntityTable from '../components/common/EntityTable';
import { GridColDef } from '@mui/x-data-grid';

export function OpsSchedulingPage() {
  const { token } = useAuth();
  const [hostId, setHostId] = useState('1');
  const [roomId, setRoomId] = useState('1');
  const [hostAgenda, setHostAgenda] = useState<Array<Record<string, unknown>>>([]);
  const [roomAgenda, setRoomAgenda] = useState<Array<Record<string, unknown>>>([]);

  async function load() {
    if (!token) return;
    const h = await api.hostAgenda(token, Number(hostId));
    const r = await api.roomAgenda(token, Number(roomId));
    setHostAgenda(h.items || []);
    setRoomAgenda(r.items || []);
  }

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Scheduling Ops" subtitle="Capacity and agenda visibility for operations/admin." />
      <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
        <TextField label="Host ID" value={hostId} onChange={(e) => setHostId(e.target.value)} />
        <TextField label="Room ID" value={roomId} onChange={(e) => setRoomId(e.target.value)} />
        <Button variant="contained" onClick={() => load()}>Load</Button>
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
