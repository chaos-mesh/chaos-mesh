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
      height: '56px',
      borderBottom: `1px solid ${theme.palette.divider}`,
    },
    title: {
      fontSize: '1rem',
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
      <Typography className={classes.title} variant="button">
        {title}
      </Typography>
      {children}
    </Box>
  )
}

export default PaperTop
