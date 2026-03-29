import { Alert, Card, CardContent, CardMedia, CircularProgress, Grid2 as Grid, Stack, Typography } from '@mui/material';
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

  const subtitle = useMemo(() => `${rows.length} published package calendar rows loaded`, [rows.length]);

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Travel Catalog" subtitle={subtitle} />
      {loading ? <CircularProgress /> : <CatalogTable rows={rows} />}
      <Grid container spacing={2}>
        {[{ title: 'Routes', items: routes }, { title: 'Partner Hotels', items: hotels }, { title: 'Attractions', items: attractions }].map((section) => (
          <Grid key={section.title} size={{ xs: 12, md: 4 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>{section.title}</Typography>
            {section.items.length === 0 ? (
              <Alert severity="info">No published {section.title.toLowerCase()}.</Alert>
            ) : (
              <Stack spacing={1.2}>
                {section.items.map((it, idx) => (
                  <Card key={idx} variant="outlined">
                    <CardMedia component="img" height="120" image={String((it.imagePaths as string[] | undefined)?.[0] || '/placeholder.jpg')} />
                    <CardContent>
                      <Typography variant="subtitle1">{String(it.name || '-')}</Typography>
                      <Typography variant="body2" color="text.secondary">{String(it.richDescription || '')}</Typography>
                    </CardContent>
                  </Card>
                ))}
              </Stack>
            )}
          </Grid>
        ))}
      </Grid>
    </Stack>
  );
}
