import { Chip, Paper } from '@mui/material';
import { DataGrid, GridColDef } from '@mui/x-data-grid';

type CatalogRow = {
  id: number;
  name: string;
  destination: string;
  inventoryRemaining: number;
  serviceDate: string;
  blackoutNote: string;
};

const columns: GridColDef<CatalogRow>[] = [
  { field: 'name', headerName: 'Package', flex: 1.2, minWidth: 200 },
  { field: 'destination', headerName: 'Destination', flex: 1, minWidth: 140 },
  { field: 'serviceDate', headerName: 'Date', width: 130 },
  {
    field: 'inventoryRemaining',
    headerName: 'Inventory',
    width: 120,
    renderCell: (params) => (
      <Chip
        size="small"
        color={params.value > 3 ? 'success' : 'warning'}
        label={String(params.value ?? 0)}
      />
    )
  },
  {
    field: 'blackoutNote',
    headerName: 'Blackout Note',
    flex: 1,
    minWidth: 220,
    renderCell: (params) => <span style={{ color: '#b45309' }}>{params.value || '-'}</span>
  }
];

export function CatalogTable({ rows }: { rows: CatalogRow[] }) {
  return (
    <Paper sx={{ height: 510, overflow: 'hidden' }}>
      <DataGrid
        rows={rows}
        columns={columns}
        disableRowSelectionOnClick
        pageSizeOptions={[10, 25, 50]}
        initialState={{ pagination: { paginationModel: { pageSize: 10, page: 0 } } }}
      />
    </Paper>
  );
}
