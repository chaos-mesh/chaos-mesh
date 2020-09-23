import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'

import { EventPod } from 'api/events.type'
import React from 'react'
import T from 'components/T'

const AffectedPods: React.FC<{ pods: EventPod[] }> = ({ pods }) => (
  <TableContainer>
    <Table size="small">
      <TableHead>
        <TableRow>
          <TableCell>{T('events.event.pod.ip')}</TableCell>
          <TableCell>{T('events.event.pod.name')}</TableCell>
          <TableCell>{T('events.event.pod.namespace')}</TableCell>
          <TableCell>{T('events.event.pod.action')}</TableCell>
          <TableCell>{T('events.event.pod.message')}</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {pods &&
          pods.map((pod, i) => (
            <TableRow key={i}>
              <TableCell>{pod.pod_ip}</TableCell>
              <TableCell>{pod.pod_name}</TableCell>
              <TableCell>{pod.namespace}</TableCell>
              <TableCell>{pod.action}</TableCell>
              <TableCell>{pod.message}</TableCell>
            </TableRow>
          ))}
      </TableBody>
    </Table>
  </TableContainer>
)

export default AffectedPods
