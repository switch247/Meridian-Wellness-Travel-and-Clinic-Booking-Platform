import OpenInNewRoundedIcon from '@mui/icons-material/OpenInNewRounded';
import { Button, Paper, Stack } from '@mui/material';
import { SectionHeader } from '../components/common/SectionHeader';

export function DocsPage() {
  return (
    <Stack spacing={2.5}>
      <SectionHeader title="API Documentation" subtitle="OpenAPI + Swagger UI served by backend /docs endpoint." />
      <Paper sx={{ p: 2.5 }}>
        <Button href="https://localhost:8443/docs" target="_blank" rel="noreferrer" variant="contained" endIcon={<OpenInNewRoundedIcon />}>
          Open Swagger Docs
        </Button>
      </Paper>
    </Stack>
  );
}
