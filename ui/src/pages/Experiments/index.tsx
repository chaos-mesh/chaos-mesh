import { Box, Button, Grid, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import ConfirmDialog from 'components-mui/ConfirmDialog'
import { Experiment } from 'api/experiments.type'
import ExperimentListItem from 'components/ExperimentListItem'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import TuneIcon from '@material-ui/icons/Tune'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

export default function Experiments() {
  const intl = useIntl()
  const history = useHistory()

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

  // Get all experiments after mount
  useEffect(fetchExperiments, [])

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
      <Box mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/newExperiment')}>
          {T('newE.title')}
        </Button>
      </Box>

      {experiments &&
        experiments.length > 0 &&
        Object.entries(_groupBy(experiments, 'kind'))
          .sort((a, b) => (a[0] > b[0] ? 1 : -1))
          .map(([kind, experimentsByKind]) => (
            <Box key={kind} mb={6}>
              <Box mb={3} ml={1}>
                <Typography variant="overline">{transByKind(kind as any)}</Typography>
              </Box>
              <Grid container spacing={6}>
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
        <NotFound textAlign="center">
          <Box mb={3}>
            <TuneIcon fontSize="large" />
          </Box>
          <Typography variant="h6">{T('experiments.noExperimentsFound')}</Typography>
        </NotFound>
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
