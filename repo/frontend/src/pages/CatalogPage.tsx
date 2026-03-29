import {
  Alert,
  Box,
  Card,
  CardContent,
  CardMedia,
  Chip,
  CircularProgress,
  Grid2 as Grid,
  Paper,
  Stack,
  Typography
} from '@mui/material';
import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { CatalogTable } from '../components/catalog/CatalogTable';
import { SectionHeader } from '../components/common/SectionHeader';

type CatalogRow = {
  id: number;
  name: string;
  destination: string;
  inventoryRemaining: number;
  serviceDate: string;
  blackoutNote: string;
};

export function CatalogPage() {
  const [rows, setRows] = useState<CatalogRow[]>([]);
  const [routes, setRoutes] = useState<Array<Record<string, unknown>>>([]);
  const [hotels, setHotels] = useState<Array<Record<string, unknown>>>([]);
  const [attractions, setAttractions] = useState<Array<Record<string, unknown>>>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([api.catalog(), api.routes(), api.hotels(), api.attractions()])
      .then(([r, rt, ht, at]) => {
        const mapped = (r.items || []).map((it, idx) => ({
          id: Number(it.id ?? idx + 1),
          name: String(it.name ?? '-'),
          destination: String(it.destination ?? '-'),
          inventoryRemaining: Number(it.inventoryRemaining ?? 0),
          serviceDate: String(it.serviceDate ?? '-').slice(0, 10),
          blackoutNote: String(it.blackoutNote ?? '')
        }));
        setRows(mapped);
        setRoutes(rt.items || []);
        setHotels(ht.items || []);
        setAttractions(at.items || []);
      })
      .finally(() => setLoading(false));
  }, []);

  const subtitle = useMemo(() => `${rows.length} published package calendars ready for booking`, [rows.length]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Travel Catalog" subtitle={subtitle} />

      <Paper sx={{ p: 2.5, backgroundImage: 'linear-gradient(135deg, #0d6e6e 0%, #2a9d8f 100%)', color: 'white' }}>
        <Typography variant="h6" gutterBottom>Inventory Snapshot</Typography>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
          <Chip label={`${rows.length} date slots`} color="warning" sx={{ color: 'white' }} />
          <Chip label={`${routes.length} guided routes`} color="success" sx={{ color: 'white' }} />
          <Chip label={`${hotels.length} partner hotels`} color="secondary" sx={{ color: 'white' }} />
          <Chip label={`${attractions.length} attractions`} color="info" sx={{ color: 'white' }} />
        </Stack>
      </Paper>

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6" gutterBottom>Published Packages</Typography>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : rows.length === 0 ? (
          <Alert severity="info">No catalog rows are published yet.</Alert>
        ) : (
          <CatalogTable rows={rows} />
        )}
      </Paper>

      <Grid container spacing={2}>
        {[
          { title: 'Routes', items: routes, accent: 'primary.main' },
          { title: 'Partner Hotels', items: hotels, accent: 'secondary.main' },
          { title: 'Attractions', items: attractions, accent: '#f97316' }
        ].map((section) => (
          <Grid key={section.title} size={{ xs: 12, md: 4 }}>
            <Paper sx={{ p: 2.5, minHeight: '100%' }}>
              <Typography variant="h6" sx={{ mb: 1 }}>{section.title}</Typography>
              {section.items.length === 0 ? (
                <Alert severity="info">No published {section.title.toLowerCase()}.</Alert>
              ) : (
                <Stack spacing={2}>
                  {section.items.map((it, idx) => (
                    <Card key={idx} variant="outlined" sx={{ borderColor: section.accent }}>
                      <CardMedia
                        component="img"
                        height="130"
                        image={String((it.imagePaths as string[] | undefined)?.[0] || '/placeholder.jpg')}
                        alt={String(it.name || '')}
                      />
                      <CardContent>
                        <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>{String(it.name || '-')}</Typography>
                        <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                          Destination ID #{String(it.destinationId ?? '-')}.
                        </Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                          {String(it.richDescription || '')}
                        </Typography>
                      </CardContent>
                    </Card>
                  ))}
                </Stack>
              )}
            </Paper>
          </Grid>
        ))}
      </Grid>
    </Stack>
  );
}
