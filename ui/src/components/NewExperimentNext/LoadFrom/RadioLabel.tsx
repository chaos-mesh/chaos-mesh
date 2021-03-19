import { Box, Typography } from '@material-ui/core'

import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'
import React from 'react'
import { truncate } from 'lib/utils'

const RadioLabel = (e: Experiment | Archive) => (
  <Box display="flex" justifyContent="space-between" alignItems="center">
    <Typography>{e.name}</Typography>
    <Box ml={3}>
      <Typography variant="body2" color="textSecondary" title={e.uid}>
        {truncate(e.uid)}
      </Typography>
    </Box>
  </Box>
)

export default RadioLabel
