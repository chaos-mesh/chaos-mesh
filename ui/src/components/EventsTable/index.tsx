import {
  Box,
  Button,
  IconButton,
  InputAdornment,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableFooter,
  TableHead,
  TablePagination,
  TableRow,
  TableSortLabel,
  TextField,
} from '@material-ui/core'
import React, { useCallback, useEffect, useState } from 'react'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import day, { dayComparator } from 'lib/dayjs'

import { Event } from 'api/events.type'
import FirstPageIcon from '@material-ui/icons/FirstPage'
import KeyboardArrowLeftIcon from '@material-ui/icons/KeyboardArrowLeft'
import KeyboardArrowRightIcon from '@material-ui/icons/KeyboardArrowRight'
import LastPageIcon from '@material-ui/icons/LastPage'
import { Link } from 'react-router-dom'
import PaperTop from 'components/PaperTop'
import SearchIcon from '@material-ui/icons/Search'
import _debounce from 'lodash.debounce'
import { searchEvents } from 'lib/search'
import { usePrevious } from 'lib/hooks'
import useRunningLabelStyles from 'lib/styles/runningLabel'

const useStyles = makeStyles(() =>
  createStyles({
    tableContainer: {
      maxHeight: 768,
    },
  })
)

function descendingComparator<T>(a: T, b: T, orderBy: keyof T) {
  if (['CreateAt', 'UpdateAt', 'StartTime', 'EndTime'].includes(orderBy as string)) {
    return dayComparator(a[orderBy] as any, b[orderBy] as any)
  }

  if (b[orderBy] < a[orderBy]) {
    return -1
  }

  if (b[orderBy] > a[orderBy]) {
    return 1
  }

  return 0
}

type Order = 'asc' | 'desc'

function getComparator<Key extends keyof any>(
  order: Order,
  orderBy: Key
): (a: { [key in Key]: number | string }, b: { [key in Key]: number | string }) => number {
  return order === 'desc'
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy)
}

function stableSort<T>(data: T[], comparator: (a: T, b: T) => number) {
  const indexed = data.map((el, index) => [el, index] as [T, number])

  indexed.sort((a, b) => {
    const order = comparator(a[0], b[0])

    if (order !== 0) {
      return order
    }

    return a[1] - b[1]
  })

  return indexed.map((el) => el[0])
}

type SortedEvent = Omit<Event, 'deleted_at' | 'pods'>
type SortedEventWithPods = Omit<Event, 'deleted_at'>

const headCells: { id: keyof SortedEvent; label: string }[] = [
  { id: 'experiment', label: 'Experiment' },
  { id: 'experiment_id', label: 'UUID' },
  { id: 'namespace', label: 'Namespace' },
  { id: 'kind', label: 'Kind' },
  { id: 'start_time', label: 'Start Time' },
  { id: 'finish_time', label: 'Finish Time' },
]

interface EventsTableHeadProps {
  order: Order
  orderBy: keyof SortedEvent
  onSort: (e: React.MouseEvent<unknown>, k: keyof SortedEvent) => void
  detailed: boolean
}

const EventsTableHead: React.FC<EventsTableHeadProps> = ({ order, orderBy, onSort, detailed }) => {
  const handleSortEvents = (k: keyof SortedEvent) => (e: React.MouseEvent<unknown>) => onSort(e, k)

  let cells = headCells
  if (detailed) {
    cells = cells.concat([{ id: 'Detail' as keyof SortedEvent, label: 'Event Detail' }])
  }

  return (
    <TableHead>
      <TableRow>
        {cells.map((cell) => (
          <TableCell
            key={cell.id}
            sortDirection={orderBy === cell.id ? order : false}
            onClick={handleSortEvents(cell.id)}
          >
            <TableSortLabel active={orderBy === cell.id} direction={orderBy === cell.id ? order : 'desc'}>
              {cell.label}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  )
}

interface TablePaginationActionsProps {
  count: number
  page: number
  rowsPerPage: number
  onChangePage: (e: React.MouseEvent<HTMLButtonElement>, newPage: number) => void
}

const TablePaginationActions: React.FC<TablePaginationActionsProps> = ({ count, page, rowsPerPage, onChangePage }) => {
  const handleFirstPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onChangePage(event, 0)
  const handleBackButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onChangePage(event, page - 1)
  const handleNextButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => onChangePage(event, page + 1)
  const handleLastPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) =>
    onChangePage(event, Math.max(0, Math.ceil(count / rowsPerPage) - 1))

  return (
    <Box display="flex" className="test-box">
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

const format = (date: string) => day(date).format('YYYY-MM-DD HH:mm:ss')

interface EventsTableRowProps {
  event: SortedEventWithPods
  detailed: boolean
}

const EventsTableRow: React.FC<EventsTableRowProps> = ({ event: e, detailed }) => {
  const runningLabel = useRunningLabelStyles()

  return (
    <>
      <TableRow hover>
        <TableCell>{e.experiment}</TableCell>
        <TableCell>{e.experiment_id}</TableCell>
        <TableCell>{e.namespace}</TableCell>
        <TableCell>{e.kind}</TableCell>
        <TableCell>{format(e.start_time)}</TableCell>
        <TableCell>
          {e.finish_time ? format(e.finish_time) : <span className={runningLabel.root}>Running</span>}
        </TableCell>
        {detailed && (
          <TableCell>
            <Button
              component={Link}
              to={`/experiments/${e.experiment_id}?name=${e.experiment}&event=${e.id}`}
              variant="outlined"
              size="small"
              color="primary"
            >
              Detail
            </Button>
          </TableCell>
        )}
      </TableRow>
    </>
  )
}

export interface EventsTableProps {
  title?: string
  events: Event[]
  detailed?: boolean
}

const EventsTable: React.FC<EventsTableProps> = ({ title = 'Events', events: allEvents, detailed = false }) => {
  const classes = useStyles()

  const [events, setEvents] = useState(allEvents)
  const [order, setOrder] = useState<Order>('desc')
  const [orderBy, setOrderBy] = useState<keyof SortedEvent>('start_time')
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(5)
  const [search, setSearch] = useState('')
  const previousSearch = usePrevious(search)

  const handleSortEvents = (_: React.MouseEvent<unknown>, k: keyof SortedEvent) => {
    const isAsc = orderBy === k && order === 'asc'

    setOrder(isAsc ? 'desc' : 'asc')
    setOrderBy(k)
  }

  const handleChangePage = (_: React.MouseEvent<HTMLButtonElement> | null, newPage: number) => setPage(newPage)

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setRowsPerPage(parseInt(event.target.value))
    setPage(0)
  }

  const debounceSetSearch = useCallback(_debounce(setSearch, 500), [])
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => debounceSetSearch(e.target.value)

  useEffect(() => {
    if (search && allEvents) {
      setEvents(searchEvents(allEvents, search))
    }

    if (previousSearch !== '' && search === '') {
      setEvents(allEvents)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search])

  return (
    <>
      <PaperTop title={title}>
        <TextField
          style={{ width: '200px', minWidth: '30%', margin: 0 }}
          margin="dense"
          placeholder="Search events ..."
          disabled={!allEvents}
          variant="outlined"
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon color="primary" />
              </InputAdornment>
            ),
          }}
          inputProps={{
            style: { paddingTop: 8, paddingBottom: 8 },
          }}
          onChange={handleSearchChange}
        />
      </PaperTop>
      <TableContainer className={classes.tableContainer}>
        <Table stickyHeader>
          <EventsTableHead order={order} orderBy={orderBy} onSort={handleSortEvents} detailed={detailed} />

          <TableBody>
            {events &&
              stableSort<SortedEvent>(events, getComparator(order, orderBy))
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((e) => <EventsTableRow key={e.id} event={e as SortedEventWithPods} detailed={detailed} />)}
          </TableBody>

          <TableFooter>
            <TableRow>
              {events && (
                <TablePagination
                  count={events.length}
                  page={page}
                  rowsPerPageOptions={[5, 10, 25]}
                  rowsPerPage={rowsPerPage}
                  onChangePage={handleChangePage}
                  onChangeRowsPerPage={handleChangeRowsPerPage}
                  ActionsComponent={TablePaginationActions as any}
                  labelDisplayedRows={({ from, to, count }) => `${from} - ${to} of ${count}`}
                  labelRowsPerPage="Events per page"
                />
              )}
            </TableRow>
          </TableFooter>
        </Table>
      </TableContainer>
    </>
  )
}

export default EventsTable
