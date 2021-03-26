import { Box, Button, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import ConfirmDialog from 'components-mui/ConfirmDialog'
import DataTable from './DataTable'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const Workflows = () => {
  const intl = useIntl()
  const history = useHistory()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'workflow',
  })

  const fetchWorkflows = () => {
    setLoading(true)

    api.workflows
      .workflows()
      .then(({ data }) => setWorkflows(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    fetchWorkflows()
  }, [])

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.experiments.deleteExperiment

        break
      default:
        actionFunc = null
    }

    if (actionFunc === null) {
      return
    }

    const { uuid } = selected

    actionFunc(uuid)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: `common.${action}Successfully` }),
          })
        )

        setTimeout(fetchWorkflows, 300)
      })
      .catch(console.error)
  }

  return (
    <>
      <Box mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/workflows/new')}>
          {T('newW.title')}
        </Button>
      </Box>

      <DataTable data={workflows} />

      {!loading && workflows.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('workflows.noWorkflowsFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        title={selected.title}
        description={selected.description}
        onConfirm={handleExperiment(selected.action)}
      />
    </>
  )
}

export default Workflows
