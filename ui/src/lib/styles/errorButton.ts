import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      color: theme.palette.error.dark,
      borderColor: theme.palette.error.dark,
    },
  })
)

export default useStyles
