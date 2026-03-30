import { Box, Button, Stack, Typography } from '@mui/material';

type SectionAction = {
  label: string;
  handler: () => void;
};

export function SectionHeader({
  title,
  subtitle,
  actions
}: {
  title: string;
  subtitle?: string;
  actions?: SectionAction[];
}) {
  return (
    <Box
      sx={{
        mb: 2.5,
        display: 'flex',
        flexWrap: 'wrap',
        gap: 2,
        alignItems: 'center',
        justifyContent: 'space-between'
      }}
    >
      <Box>
        <Typography variant="h4">{title}</Typography>
        {subtitle && (
          <Typography variant="body1" color="text.secondary" sx={{ mt: 0.5 }}>
            {subtitle}
          </Typography>
        )}
      </Box>
      {actions && actions.length > 0 && (
        <Stack direction="row" spacing={1}>
          {actions.map((action) => (
            <Button key={action.label} size="small" variant="outlined" onClick={action.handler}>
              {action.label}
            </Button>
          ))}
        </Stack>
      )}
    </Box>
  );
}
