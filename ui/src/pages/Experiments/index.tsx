import { Box, Button, Grid, Typography } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import { useEffect, useRef, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import { Experiment } from 'api/experiments.type'
import ExperimentListItem from 'components/ExperimentListItem'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import PlaylistAddCheckIcon from '@material-ui/icons/PlaylistAddCheck'
import Space from 'components-mui/Space'
import T from 'components/T'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const initialSelected = {
  uuid: '',
  title: '',
  description: '',
  action: '',
}

export default function Experiments() {
  const intl = useIntl()
  const history = useHistory()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [experiments, setExperiments] = useState<Experiment[]>([])
  const [selected, setSelected] = useState(initialSelected)
  const confirmRef = useRef<ConfirmDialogHandles>(null)

  const fetchExperiments = () => {
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

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

    confirmRef.current!.setOpen(false)

    const { uuid } = selected

    if (actionFunc) {
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
  }

  return (
    <>
      <Space mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/experiments/new')}>
          {T('newE.title')}
        </Button>
        <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={() => {}}>
          {T('common.batchOperation')}
        </Button>
      </Space>

      {experiments.length > 0 &&
        Object.entries(_groupBy(experiments, 'kind')).map(([kind, experimentsByKind]) => (
          <Box key={kind} mb={6}>
            <Box mb={3} ml={1}>
              <Typography variant="overline">{transByKind(kind as any)}</Typography>
            </Box>
            <Grid container spacing={6}>
              {experimentsByKind.length > 0 &&
                experimentsByKind.map((e) => (
                  <Grid key={e.uid} item xs={12}>
                    <ExperimentListItem experiment={e} handleSelect={setSelected} intl={intl} />
                  </Grid>
                ))}
            </Grid>
          </Box>
        ))}

      {!loading && experiments.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('experiments.noExperimentsFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        ref={confirmRef}
        title={selected.title}
        description={selected.description}
        onConfirm={handleExperiment(selected.action)}
      />
    </>
  )
}
