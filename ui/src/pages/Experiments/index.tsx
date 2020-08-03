import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { getStateofExperiments, setNeedToRefreshExperiments } from 'slices/experiments'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import ConfirmDialog from 'components/ConfirmDialog'
import { Experiment } from 'api/experiments.type'
import ExperimentPaper from 'components/ExperimentPaper'
import Loading from 'components/Loading'
import TuneIcon from '@material-ui/icons/Tune'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { dayComparator } from 'lib/dayjs'
import { toTitleCase } from 'lib/utils'
import { useSelector } from 'react-redux'

export default function Experiments() {
  const needToRefreshExperiments = useSelector((state: RootState) => state.experiments.needToRefreshExperiments)
  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'delete',
  })

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.log)
      .finally(() => setLoading(false))
  }

  const fetchEvents = (experiments: Experiment[]) => {
    api.events
      .dryEvents({ limit: 10 })
      .then(({ data }) => {
        if (data.length) {
          setExperiments(
            experiments.map((e) => {
              if (e.status === 'Failed') {
                return { ...e, events: [] }
              } else {
                const events = data
                  .filter((d) => d.experiment_id === e.uid)
                  .sort((a, b) => dayComparator(a.start_time, b.start_time))

                return {
                  ...e,
                  events: events.length > 0 ? [events[0]] : [],
                }
              }
            })
          )
        }
      })
      .catch(console.log)
  }

  // Get all experiments after mount
  useEffect(fetchExperiments, [])

  // Refresh experiments after some actions are completed
  useEffect(() => {
    if (needToRefreshExperiments) {
      fetchExperiments()
      dispatch(setNeedToRefreshExperiments(false))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [needToRefreshExperiments])

  // Refresh every experiments' events after experiments state updated
  useEffect(() => {
    if (experiments && experiments.length > 0 && !experiments[0].events) {
      fetchEvents(experiments)
    }
  }, [experiments])

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.experiments.deleteExperiment

        break
      case 'pause':
        actionFunc = api.experiments.pauseExperiment

        break
      case 'start':
        actionFunc = api.experiments.startExperiment

        break
      default:
        actionFunc = null
    }

    if (actionFunc === null) {
      return
    }

    setDialogOpen(false)

    const { uuid } = selected

    actionFunc(uuid)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: `${toTitleCase(action)}${action === 'start' ? 'ed' : 'd'} successfully!`,
          })
        )
        dispatch(setAlertOpen(true))
        dispatch(getStateofExperiments())
        fetchExperiments()
      })
      .catch(console.log)
  }

  return (
    <>
      {experiments &&
        experiments.length > 0 &&
        Object.entries(_groupBy(experiments, 'kind'))
          .sort((a, b) => (a[0] > b[0] ? 1 : -1))
          .map(([kind, experimentsByKind]) => (
            <Box key={kind} mb={6}>
              <Box mb={6}>
                <Typography variant="button">{kind}</Typography>
              </Box>
              <Grid container spacing={3}>
                {experimentsByKind.length > 0 &&
                  experimentsByKind.map((e) => (
                    <Grid key={e.uid} item xs={12}>
                      <ExperimentPaper experiment={e} handleSelect={setSelected} handleDialogOpen={setDialogOpen} />
                    </Grid>
                  ))}
              </Grid>
            </Box>
          ))}

      {!loading && experiments && experiments.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <Box mb={3}>
            <TuneIcon fontSize="large" />
          </Box>
          <Typography variant="h6" align="center">
            No experiments found. Try to create one.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={selected.title}
        description={selected.description}
        handleConfirm={handleExperiment(selected.action)}
      />
    </>
  )
}
