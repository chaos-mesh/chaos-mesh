import { Box, Button, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import DataTable from './DataTable'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { useHistory } from 'react-router-dom'

const Workflows = () => {
  const history = useHistory()

  const [loading, setLoading] = useState(false)
  const [workflows, setWorkflows] = useState<Workflow[]>([])

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
    </>
  )
}

export default Workflows
