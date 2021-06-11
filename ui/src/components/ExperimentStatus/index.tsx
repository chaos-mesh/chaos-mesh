import { Box, CircularProgress } from '@material-ui/core'

import DoneIcon from '@material-ui/icons/Done'
import React from 'react'
import { StatusOfExperiments } from 'api/experiments.type'

interface ExperimentEventsPreviewProps {
  status: keyof StatusOfExperiments
}

const ExperimentEventsPreview: React.FC<ExperimentEventsPreviewProps> = ({ status }) => {
  return status === 'finished' ? (
    <DoneIcon sx={{ color: 'success.main' }} />
  ) : status !== 'paused' ? (
    <Box display="flex" alignItems="center">
      <CircularProgress
        variant="determinate"
        size={20}
        value={100}
        sx={{ color: (theme) => theme.palette.grey[theme.palette.mode === 'light' ? 200 : 700] }}
      />
      <CircularProgress size={20} disableShrink sx={{ position: 'absolute' }} />
    </Box>
  ) : null
}

export default ExperimentEventsPreview
