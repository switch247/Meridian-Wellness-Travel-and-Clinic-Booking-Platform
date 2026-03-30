import { Alert, Box, Button, Modal, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material';
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
  const [open, setOpen] = useState(false);

  async function load() {
    if (!token) return;
    const out = await api.emailQueue(token);
    setItems(out.items || []);
  }
  useEffect(() => { load().catch(() => {}); }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Email Queue" subtitle="Internal template queue only. External providers are mocked by policy." />
      <Stack direction="row" justifyContent="flex-end">
        <Button variant="contained" onClick={() => setOpen(true)}>Queue Email</Button>
      </Stack>
      {msg && <Alert severity="success">{msg}</Alert>}
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Queued Templates</Typography>
        {items.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>No queued templates yet.</Alert>
        ) : (
          <TableContainer sx={{ mt: 1.5 }}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Template Key</TableCell>
                  <TableCell>Recipient Label</TableCell>
                  <TableCell>Status</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {items.map((i, idx) => (
                  <TableRow key={idx}>
                    <TableCell>{String(i.templateKey)}</TableCell>
                    <TableCell>{String(i.recipientLabel)}</TableCell>
                    <TableCell>{String(i.status)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>
      <Modal
        open={open}
        onClose={() => setOpen(false)}
        aria-labelledby="queue-email-modal-title"
        aria-describedby="queue-email-modal-description"
      >
        <Box sx={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', width: 500, bgcolor: 'background.paper', boxShadow: 24, p: 4 }}>
          <Typography id="queue-email-modal-title" variant="h6" component="h2" sx={{ mb: 2 }}>
            Queue Email Template
          </Typography>
          <Stack spacing={1.5} id="queue-email-modal-description">
            <TextField label="Template Key" value={templateKey} onChange={(e) => setTemplateKey(e.target.value)} />
            <TextField label="Recipient Label" value={recipientLabel} onChange={(e) => setRecipientLabel(e.target.value)} />
            <TextField label="Subject" value={subject} onChange={(e) => setSubject(e.target.value)} />
            <TextField label="Body" multiline minRows={3} value={body} onChange={(e) => setBody(e.target.value)} />
            <Stack direction="row" spacing={1.5}>
              <Button variant="contained" onClick={async () => {
                if (!token) return;
                await api.queueEmail(token, { templateKey, recipientLabel, subject, body });
                await load();
                setOpen(false);
              }}>Queue Template</Button>
              <Button variant="outlined" onClick={async () => {
                if (!token) return;
                const out = await api.exportEmailQueue(token);
                setMsg(`Exported queue: ${out.path}`);
                setOpen(false);
              }}>Export Queue</Button>
            </Stack>
          </Stack>
        </Box>
      </Modal>
    </Stack>
  );
}
