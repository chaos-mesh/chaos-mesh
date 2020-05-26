import { Box, Paper, Toolbar, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      paddingLeft: theme.spacing(3),
      paddingRight: theme.spacing(3),
    },
    toolbar: { ...theme.mixins.toolbar, justifyContent: 'space-between' },
    statusBox: {
      display: 'flex',
      marginRight: theme.spacing(4),
      '&:last-child': {
        marginRight: 0,
      },
    },
  })
)

export const CurrentStatus = () => {
  const classes = useStyles()

  return (
    <Box display="flex">
      <Box className={classes.statusBox}>
        <Typography variant="subtitle2">Total: 1</Typography>
      </Box>
      <Box className={classes.statusBox}>
        <Typography variant="subtitle2">Running: 2</Typography>
      </Box>
      <Box className={classes.statusBox}>
        <Typography variant="subtitle2">Failed: 3</Typography>
      </Box>
    </Box>
  )
}

const StatusBar: React.FC = ({ children }) => {
  const classes = useStyles()

  return (
    <Paper className={classes.root} elevation={0}>
      <Toolbar className={classes.toolbar}>
        <Box>{children}</Box>

        <CurrentStatus />
      </Toolbar>
    </Paper>
  )
}

export default StatusBar
