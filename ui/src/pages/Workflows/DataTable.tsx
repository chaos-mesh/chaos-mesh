import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'

import T from 'components/T'

const DataTable = () => {
  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>{T('common.name')}</TableCell>
            <TableCell>{T('k8s.namespace')}</TableCell>
            <TableCell>{T('workflow.partialTopology')}</TableCell>
            <TableCell>{T('workflow.time')}</TableCell>
            <TableCell>{T('workflow.state')}</TableCell>
            <TableCell>{T('workflow.created')}</TableCell>
            <TableCell>{T('common.operation')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody></TableBody>
      </Table>
    </TableContainer>
  )
}

export default DataTable
