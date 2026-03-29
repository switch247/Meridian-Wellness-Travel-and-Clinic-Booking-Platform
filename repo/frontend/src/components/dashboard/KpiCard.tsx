import { Paper, Stack, Typography } from '@mui/material';
import React from 'react';

export function KpiCard({
  icon,
  label,
  value,
  tone = 'primary'
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  tone?: 'primary' | 'secondary' | 'success';
}) {
  const colorMap: Record<string, string> = {
    primary: '#0d6e6e',
    secondary: '#d97706',
    success: '#2f855a'
  };

  return (
    <Paper sx={{ p: 2.2, border: `1px solid ${colorMap[tone]}22` }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
        <Typography variant="body2" color="text.secondary">{label}</Typography>
        {icon}
      </Stack>
      <Typography variant="h5">{value}</Typography>
    </Paper>
  );
}
