import type { DataGridProps, GridRenderCellParams } from '@mui/x-data-grid'

import { DataGrid } from '@mui/x-data-grid'
import { Typography } from '@mui/material'

export default function DataTable({ columns, sx, ...rest }: DataGridProps) {
  return (
    <DataGrid
      columns={columns.map((d) => ({
        ...d,
        flex: d.width ? undefined : 1,
        headerAlign: d.headerAlign || 'left',
        ...(d.renderCell || d.type === 'actions'
          ? {}
          : {
              renderCell: (params: GridRenderCellParams<string>) => (
                <Typography variant="body2">{params.value}</Typography>
              ),
            }),
      }))}
      sx={{
        color: 'onSurfaceVariant.main',
        '*': {
          '&:focus, &:focus-within': {
            outline: 'none !important',
          },
        },
        '.MuiDataGrid-columnHeaders': {
          bgcolor: 'surfaceVariant.main',
        },
        '.MuiDataGrid-row:hover': {
          cursor: 'pointer',
        },
        ...sx,
      }}
      autoHeight
      disableColumnMenu
      checkboxSelection={rest.rows.length > 0}
      {...rest}
    />
  )
}
