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
  Collapse,
  Typography,
} from '@material-ui/core'
import React, { useState } from 'react'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import day, { dayComparator } from 'lib/dayjs'

import { Event } from 'api/events.type'
import FirstPageIcon from '@material-ui/icons/FirstPage'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowLeftIcon from '@material-ui/icons/KeyboardArrowLeft'
import KeyboardArrowRightIcon from '@material-ui/icons/KeyboardArrowRight'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import LastPageIcon from '@material-ui/icons/LastPage'
import { Link } from 'react-router-dom'

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

type SortedEvent = Omit<Event, 'DeletedAt' | 'Pods'>
type SortedEventWithPods = Omit<Event, 'DeletedAt'>

const headCells: { id: keyof SortedEvent; label: string }[] = [
  { id: 'Experiment', label: 'Experiment' },
  { id: 'Kind', label: 'Kind' },
  { id: 'Namespace', label: 'Namespace' },
  { id: 'StartTime', label: 'Start Time' },
  { id: 'FinishTime', label: 'Finish Time' },
]

interface EventsTableHeadProps {
  order: Order
  orderBy: keyof SortedEvent
  onSort: (e: React.MouseEvent<unknown>, k: keyof SortedEvent) => void
  detailed?: boolean
}

const EventsTableHead: React.FC<EventsTableHeadProps> = ({ order, orderBy, onSort, detailed }) => {
  const handleSortEvents = (k: keyof SortedEvent) => (e: React.MouseEvent<unknown>) => onSort(e, k)

  const cells = detailed ? headCells.concat([{ id: 'Detail' as keyof SortedEvent, label: 'Event Detail' }]) : headCells

  return (
    <TableHead>
      <TableRow>
        <TableCell />
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
  detailed?: boolean
}

const EventsTableRow: React.FC<EventsTableRowProps> = ({ event: e, detailed }) => {
  const [open, setOpen] = useState(false)

  const handleToggle = () => setOpen(!open)

  return (
    <>
      <TableRow hover>
        <TableCell>
          <IconButton aria-label="Expand row" size="small" onClick={handleToggle}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell>{e.Experiment}</TableCell>
        <TableCell>{e.Kind}</TableCell>
        <TableCell>{e.Namespace}</TableCell>
        <TableCell>{format(e.StartTime)}</TableCell>
        <TableCell>{e.FinishTime ? format(e.FinishTime) : 'Not Done'}</TableCell>
        {detailed && (
          <TableCell>
            <Button
              component={Link}
              to={`/experiments/${e.Experiment}?namespace=${e.Namespace}&kind=${e.Kind}&event=${e.ID}`}
              variant="outlined"
              size="small"
              color="primary"
            >
              Detail
            </Button>
          </TableCell>
        )}
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingTop: 0, paddingBottom: 0, borderBottom: 0 }} colSpan={12}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box margin={3}>
              <Typography variant="h6" gutterBottom>
                Pods
              </Typography>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>IP</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Namespace</TableCell>
                    <TableCell>Action</TableCell>
                    <TableCell>Message</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {e.Pods &&
                    e.Pods.map((pod) => (
                      <TableRow key={pod.ID}>
                        <TableCell>{pod.PodIP}</TableCell>
                        <TableCell>{pod.PodName}</TableCell>
                        <TableCell>{pod.Namespace}</TableCell>
                        <TableCell>{pod.Action}</TableCell>
                        <TableCell>{pod.Message}</TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  )
}

interface EventsTableProps {
  events: Event[] | null
  detailed?: boolean
}

const EventsTable: React.FC<EventsTableProps> = ({ events, detailed }) => {
  const classes = useStyles()

  const [order, setOrder] = useState<Order>('desc')
  const [orderBy, setOrderBy] = useState<keyof SortedEvent>('CreatedAt')
  const [page, setPage] = React.useState(0)
  const [rowsPerPage, setRowsPerPage] = React.useState(5)

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
    <TableContainer className={classes.tableContainer}>
      <Table>
        <EventsTableHead order={order} orderBy={orderBy} onSort={handleSortEvents} detailed={detailed} />

        <TableBody>
          {events &&
            stableSort<SortedEvent>(events, getComparator(order, orderBy))
              .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
              .map((e) => <EventsTableRow key={e.ID} event={e as SortedEventWithPods} detailed={detailed} />)}
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
  )
}

export default EventsTable
