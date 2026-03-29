import { Box, Button, Card, CardContent, Stack, TextField, Typography } from '@mui/material';
import { FormEvent, useState } from 'react';

export type LoginFormValue = {
  username: string;
  password: string;
};

type Props = {
  loading: boolean;
  onSubmit: (value: LoginFormValue) => Promise<void>;
};

export function LoginCard({ loading, onSubmit }: Props) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    await onSubmit({ username, password });
  }

  return (
    <Card elevation={4} sx={{ maxWidth: 460, width: '100%' }}>
      <CardContent>
        <Typography variant="h5" sx={{ mb: 2 }}>Meridian Staff Login</Typography>
        <Box component="form" onSubmit={handleSubmit}>
          <Stack spacing={2}>
            <TextField label="Username" value={username} onChange={(e) => setUsername(e.target.value)} required />
            <TextField label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            <Button disabled={loading} type="submit" variant="contained">
              {loading ? 'Signing in...' : 'Sign in'}
            </Button>
          </Stack>
        </Box>
      </CardContent>
    </Card>
  );
}
