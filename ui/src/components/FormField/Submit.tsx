import { Box, Button, ButtonProps } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit(props: ButtonProps) {
  return (
    <Box mt={6} textAlign="right">
      <Button
        type={props.onClick ? undefined : 'submit'}
        variant="contained"
        color="primary"
        startIcon={<PublishIcon />}
        {...props}
      >
        {T('common.submit')}
      </Button>
    </Box>
  )
}
