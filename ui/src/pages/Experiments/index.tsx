import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import { Experiment } from 'api/experiments.type'
import ExperimentListItem from 'components/ExperimentListItem'
import Loading from 'components-mui/Loading'
import T from 'components/T'
import TuneIcon from '@material-ui/icons/Tune'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { dayComparator } from 'lib/dayjs'
import { transByKind } from 'lib/byKind'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

export default function Experiments() {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'archive',
  })

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  const fetchEvents = (experiments: Experiment[]) => {
    api.events
      .dryEvents()
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
      .catch(console.error)
  }

  // Get all experiments after mount
  useEffect(fetchExperiments, [])

  // Refresh every experiments' events after experiments state updated
  useEffect(() => {
    if (experiments && experiments.length > 0 && !experiments[0].events) {
      fetchEvents(experiments)
    }
  }, [experiments])

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'archive':
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
            message: intl.formatMessage({ id: `common.${action}Successfully` }),
          })
        )

        setTimeout(fetchExperiments, 300)
      })
      .catch(console.error)
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
                <Typography variant="button">{transByKind(kind as any)}</Typography>
              </Box>
              <Grid container spacing={3}>
                {experimentsByKind.length > 0 &&
                  experimentsByKind.map((e) => (
                    <Grid key={e.uid} item xs={12}>
                      <ExperimentListItem
                        experiment={e}
                        handleSelect={setSelected}
                        handleDialogOpen={setDialogOpen}
                        intl={intl}
                      />
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
            {T('experiments.noExperimentsFound')}
          </Typography>
        </Box>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={selected.title}
        description={selected.description}
        onConfirm={handleExperiment(selected.action)}
      />
    </>
  )
}
