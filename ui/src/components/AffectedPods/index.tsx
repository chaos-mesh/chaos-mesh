import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'

import { EventPod } from 'api/events.type'
import React from 'react'

const AffectedPods: React.FC<{ pods: EventPod[] }> = ({ pods }) => (
  <TableContainer>
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
        {pods &&
          pods.map((pod) => (
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
  </TableContainer>
)

export default AffectedPods
