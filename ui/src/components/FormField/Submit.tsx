import { Box, Button } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit() {
  return (
    <Box mt={6} textAlign="right">
      <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
        {T('common.submit')}
      </Button>
    </Box>
  )
}
