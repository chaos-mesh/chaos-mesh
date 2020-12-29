import { Typography } from '@material-ui/core'
import { styled } from '@material-ui/core/styles'

export default styled(Typography)({
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate3d(-50%, -50%, 0)',
})
