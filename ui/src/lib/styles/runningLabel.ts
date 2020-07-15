import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      padding: '3px 9px',
      background: theme.palette.warning.main,
      color: '#fff',
      borderRadius: 4,
      userSelect: 'none',
    },
  })
)

export default useStyles
