import { TableCell as MUITableCell, Paper, Table, TableBody, TableRow, Typography, withStyles } from '@material-ui/core'

import AffectedPods from 'components/AffectedPods'
import { Event } from 'api/events.type'
import React from 'react'
import { RootState } from 'store'
import T from 'components/T'
import { format } from 'lib/dayjs'
import useRunningLabelStyles from 'lib/styles/runningLabel'
import { useSelector } from 'react-redux'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface EventDetailProps {
  event: Event
}

const EventDetail: React.FC<EventDetailProps> = ({ event: e }) => {
  const { lang } = useSelector((state: RootState) => state.settings)

  const runningLabel = useRunningLabelStyles()

  return (
    <Table>
      <TableBody>
        <TableRow>
          <TableCell>{T('events.event.experiment')} ID</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.experiment_id}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('events.event.experiment')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.experiment}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('events.event.namespace')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.namespace}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('events.event.kind')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.kind}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('events.event.started')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {format(e.start_time, lang)}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('events.event.ended')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.finish_time ? (
                format(e.finish_time, lang)
              ) : (
                <span className={runningLabel.root}>{T('experiments.status.running')}</span>
              )}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>{T('newE.scope.affectedPods')}</TableCell>
          <TableCell>
            <Paper variant="outlined">
              <AffectedPods pods={e.pods!} />
            </Paper>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

export default EventDetail
