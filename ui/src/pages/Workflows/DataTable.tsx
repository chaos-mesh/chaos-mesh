import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'

import React from 'react'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import { useHistory } from 'react-router-dom'

interface DataTableProps {
  data: Workflow[]
}

const DataTable: React.FC<DataTableProps> = ({ data }) => {
  const history = useHistory()

  const handleJumpTo = (ns: string, name: string) => () => history.push(`/workflows/${ns}/${name}`)

  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>{T('common.name')}</TableCell>
            <TableCell>{T('workflow.entry')}</TableCell>
            <TableCell>{T('workflow.partialTopology')}</TableCell>
            <TableCell>{T('workflow.time')}</TableCell>
            <TableCell>{T('workflow.state')}</TableCell>
            <TableCell>{T('workflow.created')}</TableCell>
            <TableCell>{T('common.operation')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((d) => {
            const key = `${d.namespace}/${d.name}`

            return (
              <TableRow key={key} hover onClick={handleJumpTo(d.namespace, d.name)}>
                <TableCell>{d.name}</TableCell>
                <TableCell>{d.entry}</TableCell>
                <TableCell></TableCell>
                <TableCell></TableCell>
                <TableCell></TableCell>
                <TableCell></TableCell>
                <TableCell></TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default DataTable
