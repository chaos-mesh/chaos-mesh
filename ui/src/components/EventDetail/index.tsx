import { Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@material-ui/core'

import { Event } from 'api/events.type'
import React from 'react'
import day from 'lib/dayjs'
import useRunningLabelStyles from 'lib/styles/runningLabel'

const format = (date: string) => day(date).format('YYYY-MM-DD HH:mm:ss')

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
              {e.ExperimentID}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Experiment Name</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.Experiment}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Namespace</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.Namespace}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Kind</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.Kind}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Start Time</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {format(e.StartTime)}
            </Typography>
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Finish Time</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {e.FinishTime ? format(e.FinishTime) : <span className={runningLabel.root}>Running</span>}
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
            </Paper>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

export default EventDetail
