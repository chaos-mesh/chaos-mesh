import { Box, Typography } from '@material-ui/core'

import React from 'react'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
  root: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    width: '100%',
    height: 56,
  },
})

interface PaperTopProps {
  title?: string | JSX.Element
}

const PaperTop: React.FC<PaperTopProps> = ({ title, children }) => {
  const classes = useStyles()

  return (
    <Box className={classes.root} px={3}>
      <Typography variant="h6">{title}</Typography>
      {children}
    </Box>
  )
}

export default PaperTop
