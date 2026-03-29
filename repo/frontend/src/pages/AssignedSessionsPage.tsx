import { Alert, Stack } from '@mui/material';
import { SectionHeader } from '../components/common/SectionHeader';

export function AssignedSessionsPage() {
  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Assigned Sessions" subtitle="Detailed clinician session workflow will expand in next scheduling slice." />
      <Alert severity="info">No assigned sessions yet.</Alert>
    </Stack>
  );
}
