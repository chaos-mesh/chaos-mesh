import { styled } from '@material-ui/core/styles'

export default styled('span')(({ theme }) => ({
  display: 'inline-block',
  padding: '3px 9px',
  background: theme.palette.warning.dark,
  color: '#fff',
  borderRadius: 3,
  userSelect: 'none',
}))
