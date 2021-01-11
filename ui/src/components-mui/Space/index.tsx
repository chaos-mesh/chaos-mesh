import { Box } from '@material-ui/core'
import { styled } from '@material-ui/core/styles'

export default styled(Box)(({ theme }) => ({
  '& > *': {
    marginRight: theme.spacing(3),
    '&:last-child': {
      marginRight: 0,
    },
  },
}))
