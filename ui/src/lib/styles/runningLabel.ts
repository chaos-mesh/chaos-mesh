import { createStyles, makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) =>
  createStyles({
    root: {
      display: 'inline-block',
      padding: '3px 9px',
      background: theme.palette.warning.dark,
      color: '#fff',
      borderRadius: 4,
      userSelect: 'none',
    },
  })
)

export default useStyles
