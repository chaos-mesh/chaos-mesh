import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { IconButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import { useStoreDispatch, useStoreSelector } from 'store'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import DateTime from 'lib/luxon'
import Paper from 'components-mui/Paper'
import Space from 'components-mui/Space'
import StatusLabel from 'components-mui/StatusLabel'
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

  const handleJumpTo = (uuid: uuid) => () => history.push(`/workflows/${uuid}`)

  const handleSelect = (selected: Confirm) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    dispatch(setConfirm(selected))
  }

  const handleAction = (action: string, uuid: uuid) => () => {
    let actionFunc: any

    switch (action) {
      case 'archive':
        actionFunc = api.workflows.del

        break
      default:
        actionFunc = null
    }

    if (actionFunc) {
      actionFunc(uuid)
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
            {/* <TableCell>{T('workflow.time')}</TableCell> */}
            <TableCell>{T('common.status')}</TableCell>
            <TableCell>{T('table.created')}</TableCell>
            <TableCell>{T('common.operation')}</TableCell>
          </TableRow>
        </TableHead>

        <TableBody>
          {data.map((d) => (
            <TableRow key={d.uid} hover sx={{ cursor: 'pointer' }} onClick={handleJumpTo(d.uid)}>
              <TableCell>{d.name}</TableCell>
              {/* <TableCell></TableCell> */}
              <TableCell>
                <StatusLabel status={d.status} />
              </TableCell>
              <TableCell>
                {DateTime.fromISO(d.created_at, {
                  locale: lang,
                }).toRelative()}
              </TableCell>
              <TableCell>
                <Space direction="row">
                  <IconButton
                    color="primary"
                    title={T('archives.single', intl)}
                    component="span"
                    size="small"
                    onClick={handleSelect({
                      title: `${T('archives.single', intl)} ${d.name}`,
                      description: T('workflows.deleteDesc', intl),
                      handle: handleAction('archive', d.uid),
                    })}
                  >
                    <ArchiveOutlinedIcon />
                  </IconButton>
                </Space>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default DataTable
