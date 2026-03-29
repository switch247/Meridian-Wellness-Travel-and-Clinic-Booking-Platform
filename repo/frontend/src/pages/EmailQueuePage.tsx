import { Alert, Button, Paper, Stack, TextField, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

export function EmailQueuePage() {
  const { token } = useAuth();
  const [items, setItems] = useState<Array<Record<string, unknown>>>([]);
  const [templateKey, setTemplateKey] = useState('booking_confirmation');
  const [recipientLabel, setRecipientLabel] = useState('traveler@example.com');
  const [subject, setSubject] = useState('Booking Confirmation');
  const [body, setBody] = useState('Your reservation is ready for manual send.');
  const [msg, setMsg] = useState<string | null>(null);

  async function load() {
    if (!token) return;
    const out = await api.emailQueue(token);
    setItems(out.items || []);
  }
  useEffect(() => { load().catch(() => {}); }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Email Queue" subtitle="Internal template queue only. External providers are mocked by policy." />
      <Paper sx={{ p: 2.5 }}>
        <Stack spacing={1.5}>
          <TextField label="Template Key" value={templateKey} onChange={(e) => setTemplateKey(e.target.value)} />
          <TextField label="Recipient Label" value={recipientLabel} onChange={(e) => setRecipientLabel(e.target.value)} />
          <TextField label="Subject" value={subject} onChange={(e) => setSubject(e.target.value)} />
          <TextField label="Body" multiline minRows={3} value={body} onChange={(e) => setBody(e.target.value)} />
          <Stack direction="row" spacing={1.5}>
            <Button variant="contained" onClick={async () => {
              if (!token) return;
              await api.queueEmail(token, { templateKey, recipientLabel, subject, body });
              await load();
            }}>Queue Template</Button>
            <Button variant="outlined" onClick={async () => {
              if (!token) return;
              const out = await api.exportEmailQueue(token);
              setMsg(`Exported queue: ${out.path}`);
            }}>Export Queue</Button>
          </Stack>
        </Stack>
      </Paper>
      {msg && <Alert severity="success">{msg}</Alert>}
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Queued Templates</Typography>
        {items.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>No queued templates yet.</Alert>
        ) : (
          <Stack spacing={1.2} sx={{ mt: 1.5 }}>
            {items.map((i, idx) => (
              <Typography key={idx}>{String(i.templateKey)} | {String(i.recipientLabel)} | {String(i.status)}</Typography>
            ))}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
