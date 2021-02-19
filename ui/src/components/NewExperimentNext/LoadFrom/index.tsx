import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components-mui/SkeletonN'
import T from 'components/T'
import YAML from 'components/YAML'
import _snakecase from 'lodash.snakecase'
import api from 'api'
import { setExternalExperiment } from 'slices/experiments'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import { yamlToExperiment } from 'lib/formikhelpers'

const LoadFrom = () => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [experiments, setExperiments] = useState<Experiment[]>()
  const [archives, setArchives] = useState<Archive[]>()
  const [radio, setRadio] = useState('')

  const fetchExperiments = () =>
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.error)

  const fetchArchives = () =>
    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.error)

  useEffect(() => {
    Promise.all([fetchExperiments(), fetchArchives()])
  }, [])

  function fillExperiment(original: any) {
    const y = yamlToExperiment(original)
    const kind = y.target.kind

    dispatch(
      setExternalExperiment({
        kindAction: [kind, y.target[_snakecase(kind)].action ?? ''],
        target: y.target,
        basic: y.basic,
      })
    )
  }

  const onRadioChange = (e: any) => {
    const [type, uuid] = e.target.value.split('+')
    const apiRequest = type === 'e' ? api.experiments : api.archives

    setRadio(e.target.value)

    apiRequest
      .detail(uuid)
      .then(({ data }) => {
        fillExperiment(data.yaml)

        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.loadSuccessfully' }),
          })
        )
      })
      .catch(console.error)
  }

  return (
    <Paper>
      <PaperTop title={T('newE.loadFrom')}>
        <YAML callback={fillExperiment} />
      </PaperTop>
      <Box p={3} maxHeight={450} style={{ overflowY: 'scroll' }}>
        <RadioGroup value={radio} onChange={onRadioChange}>
          <Box mb={3}>
            <Typography>{T('experiments.title')}</Typography>
          </Box>
          {experiments && experiments.length > 0 ? (
            experiments.map((e) => (
              <FormControlLabel
                key={e.uid}
                value={`e+${e.uid}`}
                control={<Radio color="primary" />}
                label={RadioLabel(e)}
              />
            ))
          ) : experiments?.length === 0 ? (
            <Typography variant="body2">{T('experiments.noExperimentsFound')}</Typography>
          ) : (
            <SkeletonN n={3} />
          )}
          <Box my={6}>
            <Divider />
          </Box>
          <Box mb={3}>
            <Typography>{T('archives.title')}</Typography>
          </Box>
          {archives && archives.length > 0 ? (
            archives.map((a) => (
              <FormControlLabel
                key={a.uid}
                value={`a+${a.uid}`}
                control={<Radio color="primary" />}
                label={RadioLabel(a)}
              />
            ))
          ) : archives?.length === 0 ? (
            <Typography variant="body2">{T('archives.noArchivesFound')}</Typography>
          ) : (
            <SkeletonN n={3} />
          )}
        </RadioGroup>
      </Box>
    </Paper>
  )
}

export default LoadFrom
