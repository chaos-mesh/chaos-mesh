import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { RootState, useStoreDispatch } from 'store'

import ConfirmDialog from 'components/ConfirmDialog'
import ContentContainer from '../../components/ContentContainer'
import { Experiment } from 'api/experiments.type'
import ExperimentCard from 'components/ExperimentCard'
import InboxIcon from '@material-ui/icons/Inbox'
import Loading from 'components/Loading'
import api from 'api'
import day from 'lib/dayjs'
import { getStateofExperiments } from 'slices/globalStatus'
import { setNeedToRefreshExperiments } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

export default function Experiments() {
  const needToRefreshExperiments = useSelector((state: RootState) => state.globalStatus.needToRefreshExperiments)
  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)
  const [selected, setSelected] = useState({
    namespace: '',
    name: '',
    kind: '',
  })
  const [dialogOpen, setDialogOpen] = useState(false)

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  const fetchEvents = (experiments: Experiment[]) => {
    api.events
      .events()
      .then(({ data }) => {
        setExperiments(
          experiments.map((e) => {
            e.events = data
              .filter((d) => d.Experiment === e.Name)
              .sort((a, b) => {
                if (day(a.CreateAt).isBefore(b.CreateAt)) {
                  return -1
                }

                if (day(a.CreateAt).isAfter(b.CreateAt)) {
                  return 1
                }

                return 0
              })

            e.events = e.events.length > 3 ? e.events.reverse().slice(0, 3).reverse() : e.events

            return e
          })
        )
      })
      .catch(console.log)
  }

  useEffect(fetchExperiments, [])

  useEffect(() => {
    if (needToRefreshExperiments) {
      fetchExperiments()
      dispatch(setNeedToRefreshExperiments(false))
    }
  }, [dispatch, needToRefreshExperiments])

  useEffect(() => {
    if (experiments && experiments.length > 0 && !experiments[0].events) {
      fetchEvents(experiments)
    }
  }, [experiments])

  const handleDeleteExperiment = () => {
    setDialogOpen(false)

    const { namespace, name, kind } = selected

    api.experiments
      .deleteExperiment(namespace, name, kind)
      .then(() => {
        dispatch(getStateofExperiments())
        fetchExperiments()
      })
      .catch(console.log)
  }

  return (
    <ContentContainer>
      <Grid container spacing={3}>
        {experiments &&
          experiments.length > 0 &&
          experiments.map((e) => (
            <Grid key={e.Name} item xs={12} sm={12} md={6} lg={4} xl={3}>
              <ExperimentCard experiment={e} handleSelect={setSelected} handleDialogOpen={setDialogOpen} />
            </Grid>
          ))}
      </Grid>

      {!loading && experiments && experiments.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <InboxIcon fontSize="large" />
          <Typography variant="h6" align="center">
            No experiments found. Try to create one.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={`Delete ${selected.name}?`}
        description="Once you delete this experiment, it can't be recovered."
        handleConfirm={handleDeleteExperiment}
      />
    </ContentContainer>
  )
}
