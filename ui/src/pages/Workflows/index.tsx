import { Box, Button } from '@material-ui/core'
import { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import DataTable from './DataTable'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { useHistory } from 'react-router-dom'

const Workflows = () => {
  const history = useHistory()

  const [workflows, setWorkflows] = useState<Workflow[]>([])

  const fetchWorkflows = () =>
    api.workflows
      .workflows()
      .then(({ data }) => setWorkflows(data))
      .catch(console.error)

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
    </>
  )
}

export default Workflows
