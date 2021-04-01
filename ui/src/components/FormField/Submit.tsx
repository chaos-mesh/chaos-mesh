import { Box, Button } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit({ onClick }: { onClick?: () => void }) {
  return (
    <Box mt={6} textAlign="right">
      <Button
        type={onClick ? undefined : 'submit'}
        variant="contained"
        color="primary"
        startIcon={<PublishIcon />}
        onClick={onClick}
      >
        {T('common.submit')}
      </Button>
    </Box>
  )
}
