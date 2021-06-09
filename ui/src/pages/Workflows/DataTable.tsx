import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { IconButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import { useStoreDispatch, useStoreSelector } from 'store'

import DateTime from 'lib/luxon'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import Paper from 'components-mui/Paper'
import Space from 'components-mui/Space'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

interface DataTableProps {
  data: Workflow[]
  fetchData: () => void
}

const DataTable: React.FC<DataTableProps> = ({ data, fetchData }) => {
  const history = useHistory()
  const intl = useIntl()

  const { lang } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const handleJumpTo = (ns: string, name: string) => () => history.push(`/workflows/${ns}/${name}`)

  const handleSelect = (selected: Confirm) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    dispatch(setConfirm(selected))
  }

  const handleAction = (action: string, data: { namespace: string; name: string }) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.workflows.del

        break
      default:
        actionFunc = null
    }

    const { namespace, name } = data

    if (actionFunc) {
      actionFunc(namespace, name)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: T(`confirm.success.${action}`, intl),
            })
          )

          setTimeout(fetchData, 300)
        })
        .catch(console.error)
    }
  }

  return (
    <TableContainer component={(props) => <Paper {...props} sx={{ p: 0, borderBottom: 'none' }} />}>
      <Table stickyHeader>
        <TableHead>
          <TableRow>
            <TableCell>{T('common.name')}</TableCell>
            {/* <TableCell>{T('workflow.entry')}</TableCell> */}
            {/* <TableCell>{T('workflow.time')}</TableCell> */}
            <TableCell>{T('workflow.state')}</TableCell>
            <TableCell>{T('table.created')}</TableCell>
            <TableCell>{T('common.operation')}</TableCell>
          </TableRow>
        </TableHead>

        <TableBody>
          {data.map((d) => {
            const key = `${d.namespace}/${d.name}`

            return (
              <TableRow key={key} hover sx={{ cursor: 'pointer' }} onClick={handleJumpTo(d.namespace, d.name)}>
                <TableCell>{d.name}</TableCell>
                {/* <TableCell>{d.entry}</TableCell> */}
                {/* <TableCell></TableCell> */}
                <TableCell>{d.status}</TableCell>
                <TableCell>
                  {DateTime.fromISO(d.created, {
                    locale: lang,
                  }).toRelative()}
                </TableCell>
                <TableCell>
                  <Space direction="row">
                    <IconButton
                      color="primary"
                      title={T('common.delete', intl)}
                      component="span"
                      size="small"
                      onClick={handleSelect({
                        title: `${T('common.delete', intl)} ${d.name}`,
                        description: T('workflows.deleteDesc', intl),
                        handle: handleAction('delete', { namespace: d.namespace, name: d.name }),
                      })}
                    >
                      <DeleteOutlinedIcon />
                    </IconButton>
                  </Space>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default DataTable
