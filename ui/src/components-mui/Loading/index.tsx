import { Box, CircularProgress } from '@material-ui/core'

import { styled } from '@material-ui/styles'

const StyledBox = styled(Box)({
  position: 'absolute',
  top: 0,
  left: 0,
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  width: '100%',
  height: '100%',
})

const Loading = () => (
  <StyledBox>
    <CircularProgress size={25} />
  </StyledBox>
)

export default Loading
