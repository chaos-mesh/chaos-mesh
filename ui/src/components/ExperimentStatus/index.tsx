import { Box, CircularProgress } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import CloseIcon from '@material-ui/icons/Close'
import DoneIcon from '@material-ui/icons/Done'
import React from 'react'
import { StateOfExperiments } from 'api/experiments.type'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      position: 'relative',
      display: 'flex',
      alignItems: 'center',
    },
    success: {
      color: theme.palette.success.main,
    },
    error: {
      color: theme.palette.error.main,
    },
    bottom: {
      color: theme.palette.grey[theme.palette.type === 'light' ? 200 : 700],
    },
    top: {
      position: 'absolute',
      left: 0,
    },
  })
)

interface ExperimentEventsPreviewProps {
  status: keyof StateOfExperiments
}

const ExperimentEventsPreview: React.FC<ExperimentEventsPreviewProps> = ({ status }) => {
  const classes = useStyles()

  return (
    <Box className={classes.root}>
      {status !== 'Running' && status !== 'Failed' ? (
        <DoneIcon className={classes.success} />
      ) : status === 'Running' ? (
        <Box display="flex" alignItems="center">
          <CircularProgress className={classes.bottom} variant="determinate" size={20} value={100} />
          <CircularProgress className={classes.top} size={20} disableShrink />
        </Box>
      ) : (
        <CloseIcon className={classes.error} />
      )}
    </Box>
  )
}

export default ExperimentEventsPreview
