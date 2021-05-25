import AddIcon from '@material-ui/icons/Add'
import { Button } from '@material-ui/core'
import Space from 'components-mui/Space'
import T from 'components/T'
import { useHistory } from 'react-router-dom'

const Schedules = () => {
  const history = useHistory()

  return (
    <>
      <Space mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/schedules/new')}>
          {T('newS.title')}
        </Button>
      </Space>
    </>
  )
}

export default Schedules
