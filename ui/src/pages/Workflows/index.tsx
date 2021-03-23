import { Box, Button } from '@material-ui/core'

import AddIcon from '@material-ui/icons/Add'
import DataTable from './DataTable'
import T from 'components/T'
import { useHistory } from 'react-router-dom'

const Workflows = () => {
  const history = useHistory()

  return (
    <>
      <Box mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/workflows/new')}>
          {T('newW.title')}
        </Button>
      </Box>
      <DataTable />
    </>
  )
}

export default Workflows
