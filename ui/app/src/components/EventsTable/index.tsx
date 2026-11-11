/*
 * Copyright 2021 Chaos Mesh Authors.
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
import type { EventPage, EventPageParams, EventSort } from '@/api/events'
import Paper from '@/mui-extends/Paper'
import type { CoreEvent as Event } from '@/openapi/index.schemas'
import FirstPageIcon from '@mui/icons-material/FirstPage'
import KeyboardArrowLeftIcon from '@mui/icons-material/KeyboardArrowLeft'
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight'
import LastPageIcon from '@mui/icons-material/LastPage'
import {
  Box,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableFooter,
  TableHead,
  TablePagination,
  TableRow,
  TableSortLabel,
} from '@mui/material'
import _ from 'lodash'
import { useIntl } from 'react-intl'

import i18n from '@/components/T'

import { format } from '@/lib/luxon'

const headCells: { id: EventSort; label: string }[] = [
  { id: 'object_id', label: 'uuid' },
  { id: 'namespace', label: 'namespace' },
  { id: 'name', label: 'name' },
  { id: 'kind', label: 'kind' },
  { id: 'created_at', label: 'started' },
  { id: 'message', label: 'message' },
]

interface EventsTableHeadProps {
  order: EventPageParams['order']
  orderBy: EventSort
  pending: boolean
  onSort: (k: EventSort) => void
}

const Head: ReactFCWithChildren<EventsTableHeadProps> = ({ order, orderBy, pending, onSort }) => {
  return (
    <TableHead>
      <TableRow>
        {headCells.map((cell) => (
          <TableCell
            key={cell.id}
            sortDirection={orderBy === cell.id ? order : false}
            onClick={() => {
              if (!pending) onSort(cell.id)
            }}
          >
            <TableSortLabel
              active={orderBy === cell.id}
              direction={orderBy === cell.id ? order : 'desc'}
              disabled={pending}
              onClick={(event) => {
                event.stopPropagation()
                if (!pending) onSort(cell.id)
              }}
            >
              {cell.label && i18n(`events.event.${cell.label}`)}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  )
}

interface EventsTableRowProps {
  event: Event
}

const Row: ReactFCWithChildren<EventsTableRowProps> = ({ event: e }) => (
  <TableRow hover>
    <TableCell>{_.truncate(e.object_id!)}</TableCell>
    <TableCell>{e.namespace}</TableCell>
    <TableCell>{e.name}</TableCell>
    <TableCell>{e.kind}</TableCell>
    <TableCell>{format(e.created_at!)}</TableCell>
    <TableCell>{e.message}</TableCell>
  </TableRow>
)

interface TablePaginationActionsProps {
  count: number
  disabled?: boolean
  page: number
  rowsPerPage: number
  onPageChange: (e: React.MouseEvent<HTMLButtonElement>, newPage: number) => void
}

const TablePaginationActions: ReactFCWithChildren<TablePaginationActionsProps> = ({
  count,
  disabled,
  page,
  rowsPerPage,
  onPageChange,
}) => {
  const handleFirstPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onPageChange(event, 0)
  const handleBackButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onPageChange(event, page - 1)
  const handleNextButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onPageChange(event, page + 1)
  const handleLastPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) =>
    onPageChange(event, Math.max(0, Math.ceil(count / rowsPerPage) - 1))

  return (
    <Box display="flex" ml={3}>
      <IconButton onClick={handleFirstPageButtonClick} disabled={disabled || page === 0} aria-label="first page">
        <FirstPageIcon />
      </IconButton>
      <IconButton onClick={handleBackButtonClick} disabled={disabled || page === 0} aria-label="previous page">
        <KeyboardArrowLeftIcon />
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={disabled || page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="next page"
      >
        <KeyboardArrowRightIcon />
      </IconButton>
      <IconButton
        onClick={handleLastPageButtonClick}
        disabled={disabled || page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="last page"
      >
        <LastPageIcon />
      </IconButton>
    </Box>
  )
}

interface EventsTableProps {
  data: EventPage
  order: EventPageParams['order']
  orderBy: EventSort
  pending: boolean
  onPageChange: (page: number) => void
  onRowsPerPageChange: (limit: number) => void
  onSort: (sort: EventSort) => void
}

const EventsTable: ReactFCWithChildren<EventsTableProps> = ({
  data,
  order,
  orderBy,
  pending,
  onPageChange,
  onRowsPerPageChange,
  onSort,
}) => {
  const intl = useIntl()
  const page = data.page - 1
  const emptyPage = data.events.length === 0
  // Keep back navigation when concurrent deletions leave the requested page beyond the last page.
  const paginationCount = page > 0 && data.page > data.total_pages ? -1 : data.total

  return (
    <TableContainer component={Paper} sx={{ p: 0 }} aria-busy={pending}>
      <Table stickyHeader>
        <Head order={order} orderBy={orderBy} pending={pending} onSort={onSort} />

        <TableBody>
          {data.events.map((event) => (
            <Row key={event.id} event={event} />
          ))}
          {emptyPage && (
            <TableRow>
              <TableCell colSpan={headCells.length} align="center">
                {i18n('events.notFound')}
              </TableCell>
            </TableRow>
          )}
        </TableBody>

        <TableFooter>
          <TableRow>
            <TablePagination
              style={{ borderBottom: 'none' }}
              count={paginationCount}
              page={page}
              rowsPerPageOptions={[12, 25, 50]}
              rowsPerPage={data.limit}
              disabled={pending}
              onPageChange={(_, newPage) => onPageChange(newPage + 1)}
              onRowsPerPageChange={(event) => onRowsPerPageChange(Number(event.target.value))}
              ActionsComponent={TablePaginationActions}
              labelDisplayedRows={() =>
                `${emptyPage ? 0 : page * data.limit + 1} - ${emptyPage ? 0 : page * data.limit + data.events.length} of ${data.total}`
              }
              labelRowsPerPage={i18n('events.eventsPerPage', intl)}
            />
          </TableRow>
        </TableFooter>
      </Table>
    </TableContainer>
  )
}

export default EventsTable
