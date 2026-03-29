import * as React from 'react';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { Paper } from '@mui/material';

export default function EntityTable({ rows, columns, height = 480 }: {
  rows: any[];
  columns: GridColDef[];
  height?: number;
}) {
  return (
    <Paper sx={{ height, overflow: 'hidden' }}>
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
