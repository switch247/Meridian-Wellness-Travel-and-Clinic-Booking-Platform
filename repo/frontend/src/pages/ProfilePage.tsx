import { Alert, Avatar, Box, Button, Chip, Grid2 as Grid, IconButton, Modal, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from '@mui/material';
import { Delete as DeleteIcon, Person as PersonIcon, Refresh as RefreshIcon, Add as AddIcon } from '@mui/icons-material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';
import { inCoverage, normalizeAddressInput, setCoverageRegions } from '../utils/address';

type Contact = { id: number; name: string; relationship: string; phone: string };

export function ProfilePage() {
  const { me, token, refreshMe } = useAuth();
  const [form, setForm] = useState({ line1: '', line2: '', city: '', state: '', postalCode: '' });
  const [result, setResult] = useState<{ normalized?: string; duplicate?: boolean; inCoverage?: boolean } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [addressModalOpen, setAddressModalOpen] = useState(false);
  const [formErrors, setFormErrors] = useState<{ [key: string]: string }>({});

  const validateForm = () => {
    const errors: { [key: string]: string } = {};
    if (!form.line1.trim()) errors.line1 = 'Street address is required';
    if (!form.city.trim()) errors.city = 'City is required';
    if (!form.state.trim()) errors.state = 'State is required';
    if (!form.postalCode.trim()) errors.postalCode = 'ZIP code is required';
    else if (!/^\d{5}$/.test(form.postalCode.trim())) errors.postalCode = 'ZIP code must be 5 digits';
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };
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
    api.config.coverage()
      .then((res) => setCoverageRegions(res.allowedRegions || []))
      .catch(() => {});
    loadAddresses().catch(() => {});
    loadContacts().catch(() => {});
  }, [token]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Profile & Address Book" subtitle="Sensitive values are masked in UI by design." />
      <Grid container spacing={2}>
        <Grid size={{ xs: 12 }}>
          <Paper sx={{ p: 2.5, height: '100%' }}>
            <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 2 }}>
              <Avatar sx={{ bgcolor: 'primary.main', width: 56, height: 56 }}>
                <PersonIcon />
              </Avatar>
              <Box>
                <Typography variant="h6">Authenticated Profile</Typography>
                <Typography variant="body2" color="text.secondary">Your account details</Typography>
              </Box>
            </Stack>
            <Stack spacing={1.5}>
              <Box>
                <Typography variant="body2" color="text.secondary">Username</Typography>
                <Typography variant="body1" fontWeight={500}>{me?.username || '-'}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Roles</Typography>
                <Stack direction="row" spacing={0.5} flexWrap="wrap">
                  {(me?.roles || []).length > 0 ? (
                    me.roles.map((role) => (
                      <Chip key={role} label={role} size="small" color="primary" variant="outlined" />
                    ))
                  ) : (
                    <Typography variant="body2">-</Typography>
                  )}
                </Stack>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Phone</Typography>
                <Typography variant="body1">{me?.phone || '-'}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Address</Typography>
                <Typography variant="body1">{me?.address || '-'}</Typography>
              </Box>
            </Stack>
            <Box sx={{ mt: 2 }}>
              <Button
                startIcon={<RefreshIcon />}
                variant="outlined"
                size="small"
                onClick={refreshMe}
              >
                Refresh Profile
              </Button>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>Emergency & Billing Contacts</Typography>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
        <Stack spacing={2}>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
            <TextField
              label="Full Name"
              value={contactForm.name}
              onChange={(e) => setContactForm((c) => ({ ...c, name: e.target.value }))}
              required
              fullWidth
            />
            <TextField
              label="Relationship"
              value={contactForm.relationship}
              onChange={(e) => setContactForm((c) => ({ ...c, relationship: e.target.value }))}
              placeholder="e.g., Spouse, Parent, Doctor"
              fullWidth
            />
            <TextField
              label="Phone Number"
              value={contactForm.phone}
              onChange={(e) => setContactForm((c) => ({ ...c, phone: e.target.value }))}
              required
              fullWidth
            />
            <Button
              variant="contained"
              onClick={() => {
                if (!contactForm.name || !contactForm.phone || !token) return;
                setError(null);
                setSuccess(null);
                api.addContact(token, contactForm)
                  .then(() => {
                    loadContacts();
                    setSuccess('Contact added successfully!');
                  })
                  .catch((err) => setError((err as Error).message))
                  .finally(() => setContactForm({ name: '', relationship: '', phone: '' }));
              }}
              disabled={!contactForm.name || !contactForm.phone}
            >
              Add Contact
            </Button>
          </Stack>
          {contacts.length === 0 ? (
            <Alert severity="info">No contacts added yet. Add emergency and billing contacts for better service.</Alert>
          ) : (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell><strong>Name</strong></TableCell>
                    <TableCell><strong>Relationship</strong></TableCell>
                    <TableCell><strong>Phone</strong></TableCell>
                    <TableCell align="right"><strong>Actions</strong></TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {contacts.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell>{c.name}</TableCell>
                      <TableCell>{c.relationship || '-'}</TableCell>
                      <TableCell>{c.phone}</TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => {
                            if (!token) return;
                            api.deleteContact(token, c.id)
                              .then(() => loadContacts())
                              .catch((err) => setError((err as Error).message));
                          }}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Stack>
      </Paper>

      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
        <Typography variant="h6">Saved Addresses</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => {
            setAddressModalOpen(true);
            setFormErrors({});
            setError(null);
            setSuccess(null);
            setResult(null);
          }}
        >
          Add Address
        </Button>
      </Stack>
      <Paper sx={{ p: 2.5 }}>
        {addresses.length === 0 ? (
          <Alert severity="info">No addresses saved yet. Add your first frequent address above.</Alert>
        ) : (
          <Stack spacing={1.5}>
            {addresses.map((a, idx) => (
              <Paper key={idx} variant="outlined" sx={{ p: 2, bgcolor: 'grey.50' }}>
                <Typography variant="body1" fontWeight={500}>
                  {String(a.line1Masked ?? a.line1)}
                </Typography>
                {a.line2 && (
                  <Typography variant="body2" color="text.secondary">
                    {String(a.line2)}
                  </Typography>
                )}
                <Typography variant="body2">
                  {String(a.city)}, {String(a.state)} {String(a.postalCode)}
                </Typography>
              </Paper>
            ))}
          </Stack>
        )}
      </Paper>

      <Modal
        open={addressModalOpen}
        onClose={() => setAddressModalOpen(false)}
        aria-labelledby="add-address-modal-title"
        aria-describedby="add-address-modal-description"
      >
        <Box sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: { xs: '90%', sm: 600 },
          maxHeight: '90vh',
          overflow: 'auto',
          bgcolor: 'background.paper',
          boxShadow: 24,
          p: 4,
          borderRadius: 2
        }}>
          <Typography id="add-address-modal-title" variant="h6" sx={{ mb: 2 }}>
            Add Frequent Address
          </Typography>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
          {result && (
            <Alert severity={result.inCoverage ? 'success' : 'warning'} sx={{ mb: 2 }}>
              <Typography variant="body2">
                <strong>Normalized:</strong> {result.normalized}<br />
                <strong>Duplicate:</strong> {String(result.duplicate)}<br />
                <strong>In Coverage:</strong> {String(result.inCoverage)}
              </Typography>
            </Alert>
          )}
          <Stack spacing={2} id="add-address-modal-description">
            <TextField
              label="Street Address"
              value={form.line1}
              onChange={(e) => {
                setForm((f) => ({ ...f, line1: e.target.value }));
                if (formErrors.line1) setFormErrors((err) => ({ ...err, line1: '' }));
              }}
              required
              helperText={formErrors.line1 || "Primary street address"}
              error={!!formErrors.line1}
              fullWidth
            />
            <TextField
              label="Apartment/Unit (Optional)"
              value={form.line2}
              onChange={(e) => setForm((f) => ({ ...f, line2: e.target.value }))}
              fullWidth
            />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <TextField
                label="City"
                value={form.city}
                onChange={(e) => {
                  setForm((f) => ({ ...f, city: e.target.value }));
                  if (formErrors.city) setFormErrors((err) => ({ ...err, city: '' }));
                }}
                required
                helperText={formErrors.city || ""}
                error={!!formErrors.city}
                fullWidth
              />
              <TextField
                label="State"
                value={form.state}
                onChange={(e) => {
                  setForm((f) => ({ ...f, state: e.target.value }));
                  if (formErrors.state) setFormErrors((err) => ({ ...err, state: '' }));
                }}
                required
                helperText={formErrors.state || ""}
                error={!!formErrors.state}
                fullWidth
              />
            </Stack>
            <TextField
              label="ZIP Code"
              value={form.postalCode}
              onChange={(e) => {
                setForm((f) => ({ ...f, postalCode: e.target.value }));
                if (formErrors.postalCode) setFormErrors((err) => ({ ...err, postalCode: '' }));
              }}
              required
              helperText={formErrors.postalCode || "5-digit ZIP code"}
              error={!!formErrors.postalCode}
              fullWidth
            />
            <Alert severity={liveCoverage ? 'success' : 'warning'} sx={{ mt: 1 }}>
              <Typography variant="body2">
                <strong>Preview:</strong> {liveNormalized || 'Enter address to see preview'}<br />
                <strong>Status:</strong> {duplicateHint ? 'Possible duplicate' : 'New address'} • {liveCoverage ? 'In service area' : 'Outside service area'}
              </Typography>
            </Alert>
            <Stack direction="row" spacing={2} sx={{ mt: 2 }}>
              <Button
                variant="contained"
                onClick={async () => {
                  if (!validateForm()) return;
                  setError(null);
                  setSuccess(null);
                  try {
                    if (!token) throw new Error('Please login first');
                    const out = await api.addAddress(token, form) as { normalized: string; duplicate: boolean; inCoverage: boolean };
                    setResult(out);
                    await loadAddresses();
                    setForm({ line1: '', line2: '', city: '', state: '', postalCode: '' });
                    setFormErrors({});
                    setSuccess('Address added successfully!');
                    setAddressModalOpen(false);
                  } catch (err) {
                    setError((err as Error).message);
                  }
                }}
              >
                Save Address
              </Button>
              <Button variant="outlined" onClick={() => setAddressModalOpen(false)}>
                Cancel
              </Button>
            </Stack>
          </Stack>
        </Box>
      </Modal>
    </Stack>
  );
}
