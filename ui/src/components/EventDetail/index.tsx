import { TableCell as MUITableCell, Paper, Table, TableBody, TableRow, Typography, withStyles } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import AffectedPods from 'components/AffectedPods'
import { Event } from 'api/events.type'
import Loading from 'components-mui/Loading'
import { RootState } from 'store'
import RunningLabel from 'components-mui/RunningLabel'
import T from 'components/T'
import api from 'api'
import { format } from 'lib/dayjs'
import { useSelector } from 'react-redux'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface EventDetailProps {
  eventID: string
}

const EventDetail: React.FC<EventDetailProps> = ({ eventID }) => {
  const { lang } = useSelector((state: RootState) => state.settings)

  const [loading, setLoading] = useState(false)
  const [e, setEvent] = useState<Event>()

  const fetchEvent = () => {
    setLoading(true)

    api.events
      .get(eventID)
      .then(({ data }) => setEvent(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(fetchEvent, [eventID])

  return (
    <>
      {!loading && e && (
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
                    <RunningLabel>{T('experiments.state.running')}</RunningLabel>
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
      )}
      {loading && <Loading />}
    </>
  )
}

export default EventDetail
