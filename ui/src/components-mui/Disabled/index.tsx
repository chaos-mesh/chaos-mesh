import { styled } from '@material-ui/core/styles'

export default styled('div')(({ theme }) => ({
  position: 'absolute',
  top: 0,
  left: 0,
  width: '100%',
  height: '100%',
  background: theme.palette.action.disabledBackground,
  opacity: theme.palette.action.disabledOpacity,
  cursor: 'not-allowed',
}))
