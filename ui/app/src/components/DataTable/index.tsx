/*
 * Copyright 2022 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Typography } from '@mui/material'
import type { DataGridProps, GridRenderCellParams } from '@mui/x-data-grid'
import { DataGrid } from '@mui/x-data-grid'

export default function DataTable({ columns, sx, ...rest }: DataGridProps) {
  return (
    <DataGrid
      columns={columns.map((d) => ({
        ...d,
        flex: d.width ? undefined : 1,
        headerAlign: d.headerAlign || 'left',
        ...((d.renderCell || d.type === 'actions') && {
          renderCell: ({ value }: GridRenderCellParams) => <Typography variant="body2">{value}</Typography>,
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
