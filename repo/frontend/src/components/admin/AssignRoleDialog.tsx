import * as React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField, MenuItem } from '@mui/material';

const ROLES = ['traveler', 'coach', 'operations', 'admin', 'clinician'];

export default function AssignRoleDialog({ open, onClose, onSubmit, targetUserId }: {
  open: boolean;
  onClose: () => void;
  onSubmit: (role: string) => Promise<void>;
  targetUserId: number | null;
}) {
  const [role, setRole] = React.useState('traveler');
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => { if (!open) setRole('traveler'); }, [open]);

  return (
    <Dialog open={open} onClose={onClose} fullWidth>
      <DialogTitle>Assign Role</DialogTitle>
      <DialogContent>
        <TextField select fullWidth label="Role" value={role} onChange={(e) => setRole(e.target.value)}>
          {ROLES.map(r => <MenuItem key={r} value={r}>{r}</MenuItem>)}
        </TextField>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" disabled={loading || !targetUserId} onClick={async () => {
          if (!targetUserId) return;
          setLoading(true);
          try {
            await onSubmit(role);
            onClose();
          } finally { setLoading(false); }
        }}>{loading ? 'Assigning...' : 'Assign'}</Button>
      </DialogActions>
    </Dialog>
  );
}
