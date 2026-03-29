import * as React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography, Paper } from '@mui/material';

export default function DetailsDialog({ open, title, content, onClose }: {
  open: boolean;
  title?: string;
  content: any;
  onClose: () => void;
}) {
  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>{title || 'Details'}</DialogTitle>
      <DialogContent>
        <Paper sx={{ p: 1, whiteSpace: 'pre-wrap' }}>
          <Typography variant="body2">{typeof content === 'string' ? content : JSON.stringify(content, null, 2)}</Typography>
        </Paper>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
