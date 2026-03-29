import { Alert, Button, Grid2 as Grid, Paper, Stack, TextField, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import { inCoverage, normalizeAddressInput } from '../utils/address';

type Contact = { id: number; name: string; relationship: string; phone: string };

export function ProfilePage() {
  const { me, token, refreshMe } = useAuth();
  const [form, setForm] = useState({ line1: '', line2: '', city: '', state: '', postalCode: '' });
  const [result, setResult] = useState<{ normalized?: string; duplicate?: boolean; inCoverage?: boolean } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [addresses, setAddresses] = useState<Array<Record<string, unknown>>>([]);
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [contactForm, setContactForm] = useState({ name: '', relationship: '', phone: '' });
  const liveNormalized = normalizeAddressInput(form.line1, form.city, form.state, form.postalCode);
  const duplicateHint = addresses.some((a) => String(a.normalizedKey || '').toLowerCase() === liveNormalized.toLowerCase());
  const liveCoverage = inCoverage(form.postalCode);

  async function loadAddresses() {
    if (!token) return;
    const out = await api.listAddresses(token);
    setAddresses(out.items || []);
  }

  async function loadContacts() {
    if (!token) return;
    const out = await api.listContacts(token);
    const mapped = (out.items || []).map((c: any) => ({
      id: Number(c.id),
      name: String(c.name || ''),
      relationship: String(c.relationship || ''),
      phone: String(c.phoneMasked || '')
    }));
    setContacts(mapped);
  }

  useEffect(() => {
    loadAddresses().catch(() => {});
    loadContacts().catch(() => {});
  }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Profile & Address Book" subtitle="Sensitive values are masked in UI by design." />
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 5 }}>
          <Paper sx={{ p: 2.5 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Authenticated Profile</Typography>
            <Typography>Username: {me?.username || '-'}</Typography>
            <Typography>Roles: {(me?.roles || []).join(', ') || '-'}</Typography>
            <Typography>Phone: {me?.phone || '-'}</Typography>
            <Typography>Address: {me?.address || '-'}</Typography>
            <Button sx={{ mt: 1.5 }} variant="text" onClick={refreshMe}>Refresh</Button>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 7 }}>
          <Paper sx={{ p: 2.5 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>Add Frequent Address</Typography>
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
            {result && (
              <Alert severity={result.inCoverage ? 'success' : 'warning'} sx={{ mb: 2 }}>
                normalized: {result.normalized} | duplicate: {String(result.duplicate)} | inCoverage: {String(result.inCoverage)}
              </Alert>
            )}
            <Stack spacing={1.5}>
              <TextField label="Line 1" value={form.line1} onChange={(e) => setForm((f) => ({ ...f, line1: e.target.value }))} />
              <TextField label="Line 2" value={form.line2} onChange={(e) => setForm((f) => ({ ...f, line2: e.target.value }))} />
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
                <TextField label="City" fullWidth value={form.city} onChange={(e) => setForm((f) => ({ ...f, city: e.target.value }))} />
                <TextField label="State" fullWidth value={form.state} onChange={(e) => setForm((f) => ({ ...f, state: e.target.value }))} />
              </Stack>
              <TextField label="Postal Code" value={form.postalCode} onChange={(e) => setForm((f) => ({ ...f, postalCode: e.target.value }))} />
              <Alert severity={liveCoverage ? 'info' : 'warning'}>
                Normalized preview: {liveNormalized} | duplicate: {String(duplicateHint)} | inCoverage: {String(liveCoverage)}
              </Alert>
              <Button
                variant="contained"
                onClick={async () => {
                  setError(null);
                  setResult(null);
                  try {
                    if (!token) throw new Error('Please login first');
                    const out = await api.addAddress(token, form) as { normalized: string; duplicate: boolean; inCoverage: boolean };
                    setResult(out);
                    await loadAddresses();
                  } catch (err) {
                    setError((err as Error).message);
                  }
                }}
              >
                Save Address
              </Button>
            </Stack>
          </Paper>
        </Grid>
      </Grid>

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Contacts</Typography>
        <Stack spacing={1.2} sx={{ mt: 1.2 }}>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.2}>
            <TextField label="Name" fullWidth value={contactForm.name} onChange={(e) => setContactForm((c) => ({ ...c, name: e.target.value }))} />
            <TextField label="Relationship" fullWidth value={contactForm.relationship} onChange={(e) => setContactForm((c) => ({ ...c, relationship: e.target.value }))} />
            <TextField label="Phone" fullWidth value={contactForm.phone} onChange={(e) => setContactForm((c) => ({ ...c, phone: e.target.value }))} />
            <Button variant="contained" onClick={() => {
              if (!contactForm.name || !contactForm.phone || !token) return;
              api.addContact(token, contactForm)
                .then(() => loadContacts())
                .catch((err) => setError((err as Error).message))
                .finally(() => setContactForm({ name: '', relationship: '', phone: '' }));
            }}>Add</Button>
          </Stack>
          {contacts.length === 0 ? (
            <Alert severity="info">No contacts yet. Add emergency and billing contacts.</Alert>
          ) : contacts.map((c) => (
            <Stack key={c.id} direction="row" justifyContent="space-between" alignItems="center" sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, p: 1 }}>
              <Typography>{c.name} | {c.relationship} | {c.phone}</Typography>
              <Button size="small" color="error" onClick={() => {
                if (!token) return;
                api.deleteContact(token, c.id)
                  .then(() => loadContacts())
                  .catch((err) => setError((err as Error).message));
              }}>Remove</Button>
            </Stack>
          ))}
        </Stack>
      </Paper>

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Saved Addresses</Typography>
        {addresses.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>No addresses yet. Add your first frequent address.</Alert>
        ) : addresses.map((a, idx) => (
          <Typography key={idx} sx={{ mt: 1 }}>
            {String(a.line1Masked ?? a.line1)} {String(a.city)} {String(a.state)} {String(a.postalCode)}
          </Typography>
        ))}
      </Paper>
    </Stack>
  );
}
