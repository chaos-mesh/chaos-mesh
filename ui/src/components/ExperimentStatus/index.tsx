import { Box, CircularProgress } from '@material-ui/core'

import DoneIcon from '@material-ui/icons/Done'
import React from 'react'
import { StateOfExperiments } from 'api/experiments.type'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles((theme) => ({
  success: {
    color: theme.palette.success.main,
  },
  bottom: {
    color: theme.palette.grey[theme.palette.mode === 'light' ? 200 : 700],
  },
  top: {
    position: 'absolute',
    left: 0,
  },
}))

interface ExperimentEventsPreviewProps {
  status: keyof StateOfExperiments
}

const ExperimentEventsPreview: React.FC<ExperimentEventsPreviewProps> = ({ status }) => {
  const classes = useStyles()

  return status === 'finished' ? (
    <DoneIcon className={classes.success} />
  ) : status !== 'paused' ? (
    <Box display="flex" alignItems="center">
      <CircularProgress className={classes.bottom} variant="determinate" size={20} value={100} />
      <CircularProgress className={classes.top} size={20} disableShrink />
    </Box>
  ) : null
}

export default ExperimentEventsPreview
