import { Box, Button, Grow, Typography } from '@material-ui/core'

import AddIcon from '@material-ui/icons/Add'
import DataTable from './DataTable'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { useHistory } from 'react-router-dom'
import { useIntervalFetch } from 'lib/hooks'
import { useState } from 'react'

const Workflows = () => {
  const history = useHistory()

  const [loading, setLoading] = useState(true)
  const [workflows, setWorkflows] = useState<Workflow[]>([])

  const fetchWorkflows = (intervalID?: number) => {
    api.workflows
      .workflows()
      .then(({ data }) => {
        setWorkflows(data)

        if (data.every((d) => d.status === 'finished')) {
          clearInterval(intervalID)
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useIntervalFetch(fetchWorkflows)

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div>
          <Box mb={6}>
            <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/workflows/new')}>
              {T('newW.title')}
            </Button>
          </Box>

          {workflows.length > 0 && <DataTable data={workflows} fetchData={fetchWorkflows} />}
        </div>
      </Grow>

      {!loading && workflows.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('workflows.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}

export default Workflows
