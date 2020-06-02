import { Box, CircularProgress } from '@material-ui/core'
import { createStyles, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      position: 'absolute',
      top: 0,
      left: 0,
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      width: '100%',
      height: '100%',
    },
  })
)

const Loading = () => {
  const classes = useStyles()

  return (
    <Box className={classes.root}>
      <CircularProgress />
    </Box>
  )
}

export default Loading
