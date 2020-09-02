import { createStyles, makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) =>
  createStyles({
    root: {
      color: theme.palette.error.dark,
      borderColor: theme.palette.error.dark,
    },
  })
)

export default useStyles
