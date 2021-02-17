import {
  Box,
  Button,
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
} from '@material-ui/core'
import React, { useImperativeHandle, useState } from 'react'
import { dayComparator, format } from 'lib/dayjs'
import { useHistory, useLocation } from 'react-router-dom'

import CloseIcon from '@material-ui/icons/Close'
import { Event } from 'api/events.type'
import EventDetail from 'components/EventDetail'
import FirstPageIcon from '@material-ui/icons/FirstPage'
import KeyboardArrowLeftIcon from '@material-ui/icons/KeyboardArrowLeft'
import KeyboardArrowRightIcon from '@material-ui/icons/KeyboardArrowRight'
import LastPageIcon from '@material-ui/icons/LastPage'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import RunningLabel from 'components-mui/RunningLabel'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'
import { useIntl } from 'react-intl'
import { useQuery } from 'lib/hooks'

const useStyles = makeStyles({
  eventDetailPaper: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    overflowY: 'scroll',
  },
})

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

type SortedEvent = Omit<Event, 'pods'>
type SortedEventWithPods = Event

const headCells: { id: keyof SortedEvent; label: string }[] = [
  { id: 'experiment', label: 'experiment' },
  { id: 'experiment_id', label: 'uuid' },
  { id: 'namespace', label: 'namespace' },
  { id: 'kind', label: 'kind' },
  { id: 'start_time', label: 'started' },
  { id: 'finish_time', label: 'ended' },
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
    cells = cells.concat([{ id: 'Detail' as keyof SortedEvent, label: '' }])
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
              {cell.label && T(`events.event.${cell.label}`)}
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

interface EventsTableRowProps {
  event: SortedEventWithPods
  detailed: boolean
  onSelectEvent: (e: Event) => () => void
}

const EventsTableRow: React.FC<EventsTableRowProps> = ({ event: e, detailed, onSelectEvent }) => (
  <TableRow hover>
    <TableCell>{e.experiment}</TableCell>
    <TableCell>{e.experiment_id}</TableCell>
    <TableCell>{e.namespace}</TableCell>
    <TableCell>{e.kind}</TableCell>
    <TableCell>{format(e.start_time)}</TableCell>
    <TableCell>
      {e.finish_time ? format(e.finish_time) : <RunningLabel>{T('experiments.state.running')}</RunningLabel>}
    </TableCell>
    {detailed && (
      <TableCell>
        <Button variant="outlined" size="small" color="primary" onClick={onSelectEvent(e)}>
          {T('common.detail')}
        </Button>
      </TableCell>
    )}
  </TableRow>
)

export interface EventsTableHandles {
  onSelectEvent: (e: Event) => () => void
}

interface EventsTableProps {
  events: Event[]
  detailed?: boolean
}

const EventsTable: React.ForwardRefRenderFunction<EventsTableHandles, EventsTableProps> = (
  { events: allEvents, detailed = false },
  ref
) => {
  const classes = useStyles()

  const intl = useIntl()

  const query = useQuery()
  const eventID = query.get('event_id')

  const location = useLocation()
  const history = useHistory()

  const [events] = useState(allEvents)
  const [order, setOrder] = useState<Order>('desc')
  const [orderBy, setOrderBy] = useState<keyof SortedEvent>('start_time')
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(7)

  const onSelectEvent = (e: Event) => () => {
    history.push(`${location.pathname}?event_id=${e.id}`)
  }
  // Methods exposed to the parent
  useImperativeHandle(ref, () => ({
    onSelectEvent,
  }))
  const closeEventDetail = () => {
    history.push(`${location.pathname}`)
  }

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

  return (
    <Box position="relative" minHeight={600}>
      <TableContainer component={(props) => <Paper {...props} padding={false} />}>
        <Table stickyHeader>
          <EventsTableHead order={order} orderBy={orderBy} onSort={handleSortEvents} detailed={detailed} />

          <TableBody>
            {events &&
              stableSort<SortedEvent>(events, getComparator(order, orderBy))
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((e) => (
                  <EventsTableRow
                    key={e.id}
                    event={e as SortedEventWithPods}
                    detailed={detailed}
                    onSelectEvent={onSelectEvent}
                  />
                ))}
          </TableBody>

          <TableFooter>
            <TableRow>
              {events && (
                <TablePagination
                  style={{ borderBottom: 'none' }}
                  count={events.length}
                  page={page}
                  rowsPerPageOptions={[7, 15, 25]}
                  rowsPerPage={rowsPerPage}
                  onChangePage={handleChangePage}
                  onChangeRowsPerPage={handleChangeRowsPerPage}
                  ActionsComponent={TablePaginationActions as any}
                  labelDisplayedRows={({ from, to, count }) => `${from} - ${to} of ${count}`}
                  labelRowsPerPage={intl.formatMessage({ id: 'events.eventsPerPage' })}
                />
              )}
            </TableRow>
          </TableFooter>
        </Table>
      </TableContainer>
      {eventID && (
        <Paper
          className={classes.eventDetailPaper}
          style={{
            zIndex: 3, // .MuiTableCell-stickyHeader z-index: 2
          }}
        >
          <PaperTop title={T('common.detail')}>
            <IconButton color="primary" onClick={closeEventDetail}>
              <CloseIcon />
            </IconButton>
          </PaperTop>
          <EventDetail eventID={eventID} />
        </Paper>
      )}
    </Box>
  )
}

export default React.forwardRef(EventsTable)
