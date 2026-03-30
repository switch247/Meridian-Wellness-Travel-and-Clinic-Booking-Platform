import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardMedia,
  Chip,
  CircularProgress,
  Dialog,
  DialogTitle,
  Grid2 as Grid,
  Paper,
  Snackbar,
  Stack,
  Typography,
  Container,
  Avatar,
  TextField,
  InputAdornment,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Pagination,
  Tabs,
  Tab,
  ToggleButton,
  ToggleButtonGroup
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import ViewListIcon from '@mui/icons-material/ViewList';
import ClearIcon from '@mui/icons-material/Clear';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { BookingHoldForm } from '../components/booking/BookingHoldForm';
import { CatalogTable } from '../components/catalog/CatalogTable';
import { SectionHeader } from '../components/common/SectionHeader';
import { useAuth } from '../context/AuthContext';

type CatalogRow = {
  id: number;
  name: string;
  destination: string;
  inventoryRemaining: number;
  serviceDate: string;
  blackoutNote: string;
};

// Image component with placeholder and error handling
function ImageWithPlaceholder({ src, alt, height = 200, ...props }: { src?: string; alt: string; height?: number; [key: string]: any }) {
  const [imageError, setImageError] = useState(false);
  const [imageLoaded, setImageLoaded] = useState(false);

  const handleImageError = () => {
    setImageError(true);
    setImageLoaded(true);
  };

  const handleImageLoad = () => {
    setImageLoaded(true);
  };

  if (imageError || !src) {
    return (
      <Box
        sx={{
          height,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          bgcolor: 'grey.100',
          border: '1px solid',
          borderColor: 'grey.200',
          borderRadius: 1,
          ...props.sx
        }}
        {...props}
      >
        <Stack spacing={1} alignItems="center">
          <Avatar sx={{ bgcolor: 'grey.300', width: 48, height: 48 }}>
            📷
          </Avatar>
          <Typography variant="caption" color="text.secondary">
            Image not available
          </Typography>
        </Stack>
      </Box>
    );
  }

  return (
    <Box sx={{ position: 'relative', ...props.sx }} {...props}>
      {!imageLoaded && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            bgcolor: 'grey.50',
            zIndex: 1
          }}
        >
          <CircularProgress size={24} />
        </Box>
      )}
      <Box
        component="img"
        src={src}
        alt={alt}
        onError={handleImageError}
        onLoad={handleImageLoad}
        sx={{
          width: '100%',
          height,
          objectFit: 'cover',
          borderRadius: 1,
          opacity: imageLoaded ? 1 : 0,
          transition: 'opacity 0.3s ease'
        }}
      />
    </Box>
  );
}

export function CatalogPage() {
  const { token, me } = useAuth();
  const navigate = useNavigate();
  const [rows, setRows] = useState<CatalogRow[]>([]);
  const [routes, setRoutes] = useState<Array<Record<string, unknown>>>([]);
  const [hotels, setHotels] = useState<Array<Record<string, unknown>>>([]);
  const [attractions, setAttractions] = useState<Array<Record<string, unknown>>>([]);
  const [loading, setLoading] = useState(true);
  const [bookDialogOpen, setBookDialogOpen] = useState(false);
  const [selectedPackage, setSelectedPackage] = useState<CatalogRow | null>(null);
  const [hosts, setHosts] = useState<Array<{ id: number; username: string }>>([]);
  const [rooms, setRooms] = useState<Array<{ id: number; name: string }>>([]);
  const [successAlert, setSuccessAlert] = useState(false);

  // Filtering and pagination state
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedDestination, setSelectedDestination] = useState('');
  const [availabilityFilter, setAvailabilityFilter] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const itemsPerPage = 9;

  const isTraveler = me?.roles?.includes('traveler') ?? false;

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

    // Fetch hosts and rooms for booking
    if (token) {
      api.listHosts(token).then((r) => {
        const hosts = (r.items || []).map((u) => ({ id: Number(u.id), username: String(u.username) }));
        setHosts(hosts);
      });
      // Mock rooms for now
      setRooms([{ id: 1, name: 'Room A' }, { id: 2, name: 'Room B' }]);
    }
  }, [token]);

  const onBook = (row: CatalogRow) => {
    setSelectedPackage(row);
    setBookDialogOpen(true);
  };

  const packages = rows.map((r) => ({ id: r.id, name: r.name }));

  // Get unique destinations for filter
  const destinations = useMemo(() => {
    const unique = [...new Set(rows.map(r => r.destination))].filter(Boolean);
    return unique.sort();
  }, [rows]);

  // Filter and search logic
  const filteredRows = useMemo(() => {
    return rows.filter(row => {
      const matchesSearch = searchTerm === '' ||
        row.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        row.destination.toLowerCase().includes(searchTerm.toLowerCase());

      const matchesDestination = selectedDestination === '' || row.destination === selectedDestination;

      const matchesAvailability = availabilityFilter === '' ||
        (availabilityFilter === 'available' && row.inventoryRemaining > 0) ||
        (availabilityFilter === 'limited' && row.inventoryRemaining > 0 && row.inventoryRemaining <= 5) ||
        (availabilityFilter === 'soldout' && row.inventoryRemaining === 0);

      return matchesSearch && matchesDestination && matchesAvailability;
    });
  }, [rows, searchTerm, selectedDestination, availabilityFilter]);

  // Pagination logic
  const totalPages = Math.ceil(filteredRows.length / itemsPerPage);
  const paginatedRows = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredRows.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredRows, currentPage, itemsPerPage]);

  // Reset to page 1 when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm, selectedDestination, availabilityFilter]);

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'grey.50' }}>
      <Container maxWidth="lg" sx={{ py: 4 }}>
        {/* Page Header */}
        <Box sx={{ textAlign: 'center', mb: 4 }}>
          <Typography
            variant="h3"
            component="h1"
            sx={{
              fontWeight: 700,
              background: 'linear-gradient(135deg, #0d6e6e 0%, #2a9d8f 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              mb: 1
            }}
          >
            Wellness Catalog
          </Typography>
          <Typography variant="h6" color="text.secondary" sx={{ maxWidth: 600, mx: 'auto' }}>
            Discover transformative wellness retreats and healing journeys tailored for your well-being
          </Typography>
        </Box>

        {/* Search and Filter Bar */}
        <Paper
          elevation={0}
          sx={{
            p: 3,
            mb: 4,
            border: '1px solid',
            borderColor: 'grey.200',
            borderRadius: 3,
            bgcolor: 'background.paper'
          }}
        >
          <Grid container spacing={3} alignItems="center">
            {/* Search Input */}
            <Grid size={{ xs: 12, md: 4 }}>
              <TextField
                fullWidth
                placeholder="Search packages..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon />
                    </InputAdornment>
                  ),
                }}
                variant="outlined"
                size="small"
                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
              />
            </Grid>

            {/* Destination Filter */}
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <FormControl fullWidth size="small">
                <InputLabel>Destination</InputLabel>
                <Select
                  value={selectedDestination}
                  label="Destination"
                  onChange={(e) => setSelectedDestination(e.target.value)}
                  sx={{ borderRadius: 2 }}
                >
                  <MenuItem value="">
                    <em>All Destinations</em>
                  </MenuItem>
                  {destinations.map((destination) => (
                    <MenuItem key={destination} value={destination}>
                      {destination}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            {/* Availability Filter */}
            <Grid size={{ xs: 12, sm: 6, md: 3 }}>
              <FormControl fullWidth size="small">
                <InputLabel>Availability</InputLabel>
                <Select
                  value={availabilityFilter}
                  label="Availability"
                  onChange={(e) => setAvailabilityFilter(e.target.value)}
                  sx={{ borderRadius: 2 }}
                >
                  <MenuItem value="">
                    <em>All Packages</em>
                  </MenuItem>
                  <MenuItem value="available">Available Now</MenuItem>
                  <MenuItem value="limited">Limited Spots</MenuItem>
                  <MenuItem value="booked">Fully Booked</MenuItem>
                </Select>
              </FormControl>
            </Grid>

            {/* View Mode Toggle */}
            <Grid size={{ xs: 12, md: 2 }}>
              <ToggleButtonGroup
                value={viewMode}
                exclusive
                onChange={(e, newView) => newView && setViewMode(newView)}
                size="small"
                fullWidth
                sx={{ borderRadius: 2 }}
              >
                <ToggleButton value="grid" aria-label="grid view">
                  <ViewModuleIcon />
                </ToggleButton>
                <ToggleButton value="list" aria-label="list view">
                  <ViewListIcon />
                </ToggleButton>
              </ToggleButtonGroup>
            </Grid>
          </Grid>

          {/* Filter Summary and Clear */}
          <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Showing {paginatedRows.length} of {filteredRows.length} packages
              {filteredRows.length !== rows.length && (
                <span> (filtered from {rows.length} total)</span>
              )}
            </Typography>
            {(searchTerm || selectedDestination || availabilityFilter) && (
              <Button
                size="small"
                onClick={() => {
                  setSearchTerm('');
                  setSelectedDestination('');
                  setAvailabilityFilter('');
                }}
                startIcon={<ClearIcon />}
                sx={{ borderRadius: 2 }}
              >
                Clear Filters
              </Button>
            )}
          </Box>
        </Paper>

        {/* Packages Grid */}
        {loading ? (
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 12 }}>
            <CircularProgress size={64} sx={{ mb: 3 }} />
            <Typography variant="h6" color="text.secondary">
              Loading wellness packages...
            </Typography>
          </Box>
        ) : rows.length === 0 ? (
          <Paper
            sx={{
              p: 8,
              textAlign: 'center',
              bgcolor: 'background.paper',
              border: '2px dashed',
              borderColor: 'grey.300',
              borderRadius: 3
            }}
          >
            <Avatar sx={{ width: 80, height: 80, bgcolor: 'grey.300', mx: 'auto', mb: 3 }}>
              📦
            </Avatar>
            <Typography variant="h5" color="text.secondary" gutterBottom sx={{ fontWeight: 600 }}>
              No packages available yet
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 3, maxWidth: 400, mx: 'auto' }}>
              Check back soon for our curated wellness retreats and healing journeys. We're working on bringing you transformative experiences.
            </Typography>
          </Paper>
        ) : (
          <Box>
            <Grid container spacing={3}>
              {paginatedRows.map((row) => (
                <Grid key={row.id} size={{ xs: 12, sm: 6, md: 4 }}>
                  <Card
                    sx={{
                      height: '100%',
                      display: 'flex',
                      flexDirection: 'column',
                      transition: 'all 0.3s ease',
                      cursor: isTraveler ? 'pointer' : 'default',
                      border: '1px solid',
                      borderColor: 'grey.200',
                      borderRadius: 3,
                      boxShadow: 'none',
                      '&:hover': isTraveler ? {
                        transform: 'translateY(-4px)',
                        boxShadow: '0 8px 25px rgba(0,0,0,0.1)',
                        borderColor: 'primary.light'
                      } : {}
                    }}
                    onClick={isTraveler ? () => onBook(row) : undefined}
                  >
                    <Box
                      sx={{
                        height: 160,
                        background: 'linear-gradient(135deg, #0d6e6e 0%, #2a9d8f 100%)',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        position: 'relative',
                        overflow: 'hidden',
                        borderRadius: '12px 12px 0 0'
                      }}
                    >
                      <Box
                        sx={{
                          position: 'absolute',
                          top: 0,
                          left: 0,
                          right: 0,
                          bottom: 0,
                          background: 'url("data:image/svg+xml,%3Csvg width="40" height="40" viewBox="0 0 40 40" xmlns="http://www.w3.org/2000/svg"%3E%3Cg fill="%23ffffff" fill-opacity="0.05"%3E%3Cpath d="M20 20c0-5.5-4.5-10-10-10s-10 4.5-10 10 4.5 10 10 10 10-4.5 10-10zm10 0c0-5.5-4.5-10-10-10s-10 4.5-10 10 4.5 10 10 10 10-4.5 10-10z"/%3E%3C/g%3E%3C/svg%3E")',
                          opacity: 0.3
                        }}
                      />
                      <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)', width: 56, height: 56, mb: 1 }}>
                        🏔️
                      </Avatar>
                      <Typography variant="h6" sx={{ color: 'white', fontWeight: 600, textAlign: 'center', px: 2 }}>
                        {row.name}
                      </Typography>
                    </Box>
                    <CardContent sx={{ flexGrow: 1, p: 3 }}>
                      <Stack spacing={2}>
                        <Box>
                          <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1, color: 'primary.main' }}>
                            📍 {row.destination}
                          </Typography>
                          <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            📅 {row.serviceDate}
                          </Typography>
                        </Box>

                        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                          <Chip
                            label={`${row.inventoryRemaining} spots left`}
                            size="small"
                            color={row.inventoryRemaining > 5 ? "success" : row.inventoryRemaining > 0 ? "warning" : "error"}
                            variant="outlined"
                            sx={{ borderRadius: 2 }}
                          />
                          {row.blackoutNote && (
                            <Chip
                              label="⚠️ Special conditions"
                              size="small"
                              color="info"
                              variant="outlined"
                              sx={{ borderRadius: 2 }}
                            />
                          )}
                        </Stack>

                        {isTraveler ? (
                          <Button
                            variant="contained"
                            fullWidth
                            size="large"
                            sx={{
                              mt: 'auto',
                              py: 1.5,
                              background: 'linear-gradient(135deg, #0d6e6e 0%, #2a9d8f 100%)',
                              borderRadius: 2,
                              fontWeight: 600,
                              '&:hover': {
                                background: 'linear-gradient(135deg, #1f7a8c 0%, #0d6e6e 100%)',
                                transform: 'translateY(-1px)',
                                boxShadow: '0 6px 20px rgba(13, 110, 110, 0.4)'
                              }
                            }}
                            onClick={(e) => {
                              e.stopPropagation();
                              onBook(row);
                            }}
                          >
                            Book Now →
                          </Button>
                        ) : (
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 'auto', textAlign: 'center', py: 1.5 }}>
                            Booking restricted to travelers
                          </Typography>
                        )}
                      </Stack>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>

            {/* Pagination */}
            {totalPages > 1 && (
              <Box sx={{ display: 'flex', justifyContent: 'center', mt: 6 }}>
                <Pagination
                  count={totalPages}
                  page={currentPage}
                  onChange={(e, page) => setCurrentPage(page)}
                  color="primary"
                  size="large"
                  showFirstButton
                  showLastButton
                  sx={{
                    '& .MuiPaginationItem-root': {
                      fontWeight: 500,
                      borderRadius: 2
                    }
                  }}
                />
              </Box>
            )}
          </Box>
        )}
      </Container>

      <Dialog
        open={bookDialogOpen}
        onClose={() => setBookDialogOpen(false)}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: 3,
            boxShadow: '0 24px 48px rgba(0,0,0,0.2)'
          }
        }}
      >
        <DialogTitle
          sx={{
            pb: 2,
            background: 'linear-gradient(135deg, #0d6e6e 0%, #2a9d8f 100%)',
            color: 'white',
            mb: 0
          }}
        >
          <Stack spacing={1}>
            <Typography variant="h6" component="div" sx={{ fontWeight: 600 }}>
              🌟 Book Your Wellness Journey
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.9 }}>
              {selectedPackage?.name} • {selectedPackage?.destination} • {selectedPackage?.serviceDate}
            </Typography>
          </Stack>
        </DialogTitle>
        <BookingHoldForm
          packages={packages}
          hosts={hosts}
          rooms={rooms}
          fetchSlots={async (input) => {
            if (!token) return [];
            const out = await api.availableSlots(token, input);
            return (out.items || []).map((i) => ({ slotStart: String(i.slotStart) }));
          }}
          onSubmit={async (payload) => {
            if (!token) throw new Error('Please login first');
            await api.placeHold(token, payload);
            setBookDialogOpen(false);
            setSuccessAlert(true);
            // Redirect to reservations page after a short delay
            setTimeout(() => {
              navigate('/my-reservations');
            }, 2000);
          }}
        />
      </Dialog>

      <Snackbar
        open={successAlert}
        autoHideDuration={6000}
        onClose={() => setSuccessAlert(false)}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSuccessAlert(false)}
          severity="success"
          sx={{
            width: '100%',
            boxShadow: '0 8px 32px rgba(76, 175, 80, 0.3)',
            borderRadius: 3
          }}
          icon={<span style={{ fontSize: '1.5rem' }}>🎉</span>}
        >
          <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.5 }}>
            Reservation Confirmed!
          </Typography>
          <Typography variant="body2">
            Your wellness journey awaits. Redirecting to your reservations...
          </Typography>
        </Alert>
      </Snackbar>
    </Box>
  );
}
