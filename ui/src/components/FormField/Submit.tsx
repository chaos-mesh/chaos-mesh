import { Box, Button, ButtonProps } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit({ mt = 6, ...rest }: ButtonProps & { mt?: number }) {
  return (
    <Box mt={mt} textAlign="right">
      <Button
        type={rest.onClick ? undefined : 'submit'}
        variant="contained"
        color="primary"
        startIcon={<PublishIcon />}
        {...rest}
      >
        {T('common.submit')}
      </Button>
    </Box>
  )
}
