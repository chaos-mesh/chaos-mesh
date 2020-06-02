import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { RootState, useStoreDispatch } from 'store'

import ContentContainer from '../../components/ContentContainer'
import { Experiment } from 'api/experiments.type'
import ExperimentCard from 'components/ExperimentCard'
import InboxIcon from '@material-ui/icons/Inbox'
import Loading from 'components/Loading'
import api from 'api'
import { toggleNeedToRefreshExperiments } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

export default function Experiments() {
  const needToRefreshExperiments = useSelector((state: RootState) => state.globalStatus.needToRefreshExperiments)
  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then((resp) => {
        setExperiments(resp.data)
      })
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(fetchExperiments, [])

  useEffect(() => {
    if (needToRefreshExperiments) {
      fetchExperiments()
      dispatch(toggleNeedToRefreshExperiments())
    }
  }, [dispatch, needToRefreshExperiments])

  return (
    <ContentContainer>
      <Grid container spacing={3}>
        {experiments &&
          experiments.length > 0 &&
          experiments.map((e) => (
            <Grid key={e.Name} item xs={12} sm={12} md={6} lg={4} xl={3}>
              <ExperimentCard experiment={e} />
            </Grid>
          ))}
      </Grid>

      {!loading && experiments && experiments.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <InboxIcon fontSize="large" />
          <Typography variant="h6">No experiments found. Try to create one.</Typography>
        </Box>
      )}

      {loading && <Loading />}
    </ContentContainer>
  )
}
