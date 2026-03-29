import { Alert, Stack } from '@mui/material';
import { SectionHeader } from '../components/common/SectionHeader';

export function NotFoundPage() {
  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Page Not Found" subtitle="This route does not exist in Meridian frontend." />
      <Alert severity="info">Use the left navigation to continue.</Alert>
    </Stack>
  );
}
