import { Box, Grid } from '@material-ui/core'

import LoadFrom from './LoadFrom'
import React from 'react'
import Step1 from './Step1'
import Step2 from './Step2'

const NewExperiment = () => {
  return (
    <Grid container spacing={6}>
      <Grid item xs={12} md={8}>
        <Step1 />
        <Box mt={6}>
          <Step2 />
        </Box>
      </Grid>
      <Grid item xs={12} md={4}>
        <LoadFrom from="experiments" />
        <Box mt={6}>
          <LoadFrom from="archives" />
        </Box>
        <Box mt={6}>
          <LoadFrom from="yaml" />
        </Box>
      </Grid>
    </Grid>
  )
}

export default NewExperiment
