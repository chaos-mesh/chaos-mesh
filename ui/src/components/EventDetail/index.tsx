import { TableCell as MUITableCell, Table, TableBody, TableRow, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Event } from 'api/events.type'
import Loading from 'components-mui/Loading'
import { StateOfExperimentsEnum } from 'api/experiments.type'
import T from 'components/T'
import api from 'api'
import { format } from 'lib/luxon'
import { useStoreSelector } from 'store'
import { withStyles } from '@material-ui/styles'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface EventDetailProps {
  eventID: string
}

const EventDetail: React.FC<EventDetailProps> = ({ eventID }) => {
  const { lang } = useStoreSelector((state) => state.settings)

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
              <TableCell>{T('experiments.single')} ID</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.object_id}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('experiments.single')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {/* {e.experiment} */}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('k8s.namespace')}</TableCell>
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
                  {/* {format(e.start_time, lang)} */}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('events.event.ended')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {/* {e.finish_time ? (
                    format(e.finish_time, lang)
                  ) : (
                    <StateLabel state={StateOfExperimentsEnum.Running}>{T('experiments.state.running')}</StateLabel>
                  )} */}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('newE.scope.affectedPods')}</TableCell>
              <TableCell>{/* <AffectedPods pods={e.pods!} /> */}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )}
      {loading && <Loading />}
    </>
  )
}

export default EventDetail
