import { Box, CircularProgress } from '@material-ui/core'

import React from 'react'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
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

const Loading = () => {
  const classes = useStyles()

  return (
    <Box className={classes.root}>
      <CircularProgress size={25} />
    </Box>
  )
}

export default Loading
