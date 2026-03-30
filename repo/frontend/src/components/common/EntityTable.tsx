import * as React from 'react';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { Paper } from '@mui/material';

export default function EntityTable({ rows, columns, height = 480, sx }: {
  rows: any[];
  columns: GridColDef[];
  height?: number;
  sx?: any;
}) {
  return (
    <Paper sx={{ height, overflow: 'hidden', ...sx }}>
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
