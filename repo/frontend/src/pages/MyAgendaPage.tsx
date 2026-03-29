import { Alert, Button, Paper, Stack, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import EntityTable from '../components/common/EntityTable';
import { GridColDef } from '@mui/x-data-grid';

export function MyAgendaPage() {
  const { token, me } = useAuth();
  const [items, setItems] = useState<Array<Record<string, unknown>>>([]);

  async function load() {
    if (!token || !me) return;
    const out = await api.hostAgenda(token, me.id);
    setItems(out.items || []);
  }

  useEffect(() => { load().catch(() => {}); }, [token, me?.id]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="My Agenda" subtitle="Coach/clinician assigned schedule." />
      <Button variant="outlined" onClick={() => load()}>Refresh</Button>
      <Paper sx={{ p: 2.5 }}>
        {items.length === 0 ? (
          <Alert severity="info">No assigned sessions yet.</Alert>
        ) : (
          <EntityTable
            rows={(items as any[]).map((it) => ({ id: Number(it.id), travelerId: it.travelerId, slotStart: it.slotStart, status: it.status }))}
            columns={[
              { field: 'id', headerName: 'ID', width: 90 },
              { field: 'travelerId', headerName: 'Traveler', width: 140 },
              { field: 'slotStart', headerName: 'Start', width: 220 },
              { field: 'status', headerName: 'Status', width: 140 }
            ] as GridColDef[]}
            height={380}
          />
        )}
      </Paper>
    </Stack>
  );
}
