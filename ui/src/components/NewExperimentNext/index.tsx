import { Box, Grid } from '@material-ui/core'

import LoadFromArchives from './LoadFrom/Archives'
import LoadFromExperiments from './LoadFrom/Experiments'
import LoadFromYAML from './LoadFrom/YAML'
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
        <LoadFromExperiments />
        <Box mt={6}>
          <LoadFromArchives />
        </Box>
        <Box mt={6}>
          <LoadFromYAML />
        </Box>
      </Grid>
    </Grid>
  )
}

export default NewExperiment
