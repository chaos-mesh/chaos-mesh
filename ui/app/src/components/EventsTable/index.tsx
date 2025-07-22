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
import Paper from '@/mui-extends/Paper'
import { CoreEvent as Event } from '@/openapi/index.schemas'
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
import { useState } from 'react'
import { useIntl } from 'react-intl'

import i18n from '@/components/T'

import { comparator, format } from '@/lib/luxon'

function descendingComparator<T extends Record<string, any>>(a: T, b: T, orderBy: string) {
  if (['StartTime', 'EndTime'].includes(orderBy)) {
    return comparator(a[orderBy], b[orderBy])
  }

  if (a[orderBy] > b[orderBy]) {
    return -1
  }

  if (a[orderBy] < b[orderBy]) {
    return 1
  }

  return 0
}

type Order = 'asc' | 'desc'

function getComparator<T extends Record<string, any>>(order: Order, orderBy: string): (a: T, b: T) => number {
  return order === 'desc'
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy)
}

function stableSort<T>(data: T[], comparator: (a: T, b: T) => number) {
  const indexed: [T, number][] = data.map((el, index) => [el, index])

  indexed.sort((a, b) => {
    const order = comparator(a[0], b[0])

    if (order !== 0) {
      return order
    }

    return a[1] - b[1] // compare index
  })

  return indexed.map((el) => el[0])
}

type SortedEvent = Omit<Event, 'pods'>
type SortedEventWithPods = Event

const headCells: { id: keyof SortedEvent; label: string }[] = [
  { id: 'object_id', label: 'uuid' },
  { id: 'namespace', label: 'namespace' },
  { id: 'name', label: 'name' },
  { id: 'kind', label: 'kind' },
  { id: 'created_at', label: 'started' },
  { id: 'message', label: 'message' },
]

interface EventsTableHeadProps {
  order: Order
  orderBy: keyof SortedEvent
  onSort: (e: React.MouseEvent<unknown>, k: keyof SortedEvent) => void
}

const Head: ReactFCWithChildren<EventsTableHeadProps> = ({ order, orderBy, onSort }) => {
  const handleSortEvents = (k: keyof SortedEvent) => (e: React.MouseEvent<unknown>) => onSort(e, k)

  return (
    <TableHead>
      <TableRow>
        {headCells.map((cell) => (
          <TableCell
            key={cell.id}
            sortDirection={orderBy === cell.id ? order : false}
            onClick={handleSortEvents(cell.id)}
          >
            <TableSortLabel active={orderBy === cell.id} direction={orderBy === cell.id ? order : 'desc'}>
              {cell.label && i18n(`events.event.${cell.label}`)}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  )
}

interface EventsTableRowProps {
  event: SortedEventWithPods
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
  page: number
  rowsPerPage: number
  onPageChange: (e: React.MouseEvent<HTMLButtonElement>, newPage: number) => void
}

const TablePaginationActions: ReactFCWithChildren<TablePaginationActionsProps> = ({
  count,
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
      <IconButton onClick={handleFirstPageButtonClick} disabled={page === 0} aria-label="first page">
        <FirstPageIcon />
      </IconButton>
      <IconButton onClick={handleBackButtonClick} disabled={page === 0} aria-label="previous page">
        <KeyboardArrowLeftIcon />
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="next page"
      >
        <KeyboardArrowRightIcon />
      </IconButton>
      <IconButton
        onClick={handleLastPageButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="last page"
      >
        <LastPageIcon />
      </IconButton>
    </Box>
  )
}

interface EventsTableProps {
  events: Event[]
}

const EventsTable: ReactFCWithChildren<EventsTableProps> = ({ events: allEvents }) => {
  const intl = useIntl()

  const [events] = useState(allEvents)
  const [order, setOrder] = useState<Order>('desc')
  const [orderBy, setOrderBy] = useState<keyof SortedEvent>('created_at')
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(12)

  const handleSortEvents = (_: React.MouseEvent<unknown>, k: keyof SortedEvent) => {
    const isAsc = orderBy === k && order === 'asc'

    setOrder(isAsc ? 'desc' : 'asc')
    setOrderBy(k)
  }

  const handlePageChange = (_: React.MouseEvent<HTMLButtonElement> | null, newPage: number) => setPage(newPage)

  const handleRowsPerPageChange = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setRowsPerPage(parseInt(event.target.value))
    setPage(0)
  }

  return (
    <TableContainer component={(props) => <Paper {...props} sx={{ p: 0 }} />}>
      <Table stickyHeader>
        <Head order={order} orderBy={orderBy} onSort={handleSortEvents} />

        <TableBody>
          {events &&
            stableSort<SortedEvent>(events, getComparator(order, orderBy))
              .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
              .map((e) => <Row key={e.id} event={e as SortedEventWithPods} />)}
        </TableBody>

        <TableFooter>
          <TableRow>
            {events && (
              <TablePagination
                style={{ borderBottom: 'none' }}
                count={events.length}
                page={page}
                rowsPerPageOptions={[12, 25, 50]}
                rowsPerPage={rowsPerPage}
                onPageChange={handlePageChange}
                onRowsPerPageChange={handleRowsPerPageChange}
                ActionsComponent={TablePaginationActions as any}
                labelDisplayedRows={({ from, to, count }) => `${from} - ${to} of ${count}`}
                labelRowsPerPage={i18n('events.eventsPerPage', intl)}
              />
            )}
          </TableRow>
        </TableFooter>
      </Table>
    </TableContainer>
  )
}

export default EventsTable
