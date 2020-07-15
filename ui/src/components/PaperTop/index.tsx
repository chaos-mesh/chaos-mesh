import { Box, Typography } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      width: '100%',
      height: '64px',
      borderBottom: `1px solid ${theme.palette.divider}`,
    },
  })
)

interface PaperTopProps {
  title: string
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
