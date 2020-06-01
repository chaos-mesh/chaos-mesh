import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import ContentContainer from '../../components/ContentContainer'
import { Experiment } from 'api/experiments.type'
import ExperimentCard from 'components/ExperimentCard'
import InboxIcon from '@material-ui/icons/Inbox'
import Loading from 'components/Loading'
import api from 'api'

export default function Experiments() {
  const [loading, setLoading] = useState(true)
  const [experiments, setExperiments] = useState<Experiment[]>([])

  useEffect(() => {
    api.experiments
      .experiments()
      .then((resp) => {
        setLoading(false)
        setExperiments(resp.data)
      })
      .catch(console.log)
  }, [])

  return (
    <ContentContainer>
      <Grid container spacing={3}>
        {experiments.length > 0 &&
          experiments.map((e) => (
            <Grid key={e.Name} item xs={12} sm={3}>
              <ExperimentCard experiment={e} />
            </Grid>
          ))}
      </Grid>

      {!loading && experiments.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <InboxIcon fontSize="large" />
          <Typography variant="h6">No experiments found. Try to create one.</Typography>
        </Box>
      )}

      {loading && <Loading />}
    </ContentContainer>
  )
}
