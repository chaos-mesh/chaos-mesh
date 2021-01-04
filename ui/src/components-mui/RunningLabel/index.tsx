import { styled } from '@material-ui/core/styles'

export default styled('span')(({ theme }) => ({
  display: 'inline-block',
  padding: '3px 9px',
  background: theme.palette.warning.main,
  color: theme.palette.common.white,
  borderRadius: 4,
  userSelect: 'none',
}))
