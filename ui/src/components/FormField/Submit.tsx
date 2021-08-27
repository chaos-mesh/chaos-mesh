import { Box, Button, ButtonProps } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit({ mt = 6, onClick, ...rest }: ButtonProps & { mt?: number }) {
  return (
    <Box mt={mt} textAlign="right">
      <Button
        type={onClick ? undefined : 'submit'}
        variant="contained"
        startIcon={<PublishIcon />}
        onClick={onClick}
        {...rest}
      >
        {T('common.submit')}
      </Button>
    </Box>
  )
}
