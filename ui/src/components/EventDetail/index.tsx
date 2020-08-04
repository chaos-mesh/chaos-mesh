import {
  TableCell as MUITableCell,
  Paper,
  Table,
  TableBody,
  TableHead,
  TableRow,
  Typography,
  withStyles,
} from '@material-ui/core'

import { Event } from 'api/events.type'
import React from 'react'
import { format } from 'lib/dayjs'
import useRunningLabelStyles from 'lib/styles/runningLabel'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface EventDetailProps {
  event: Event
}

const EventDetail: React.FC<EventDetailProps> = ({ event: e }) => {
  const runningLabel = useRunningLabelStyles()

  return (
    <Table>
      <TableBody>
        <TableRow>
          <TableCell>Experiment ID</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.experiment_id}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Experiment Name</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.experiment}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Namespace</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.namespace}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Kind</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.kind}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Start Time</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {format(e.start_time)}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Finish Time</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.finish_time ? format(e.finish_time) : <span className={runningLabel.root}>Running</span>}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Affected Pods</TableCell>
          <TableCell>
            <Paper variant="outlined">
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
                  {e.pods &&
                    e.pods.map((pod) => (
                      <TableRow key={pod.id}>
                        <TableCell>{pod.pod_ip}</TableCell>
                        <TableCell>{pod.pod_name}</TableCell>
                        <TableCell>{pod.namespace}</TableCell>
                        <TableCell>{pod.action}</TableCell>
                        <TableCell>{pod.message}</TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </Paper>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

export default EventDetail
