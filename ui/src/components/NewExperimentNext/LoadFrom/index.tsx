import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import { PreDefinedValue, getDB } from 'lib/idb'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'
import Paper from 'components-mui/Paper'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components-mui/SkeletonN'
import T from 'components/T'
import YAML from 'components/YAML'
import _snakecase from 'lodash.snakecase'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { setExternalExperiment } from 'slices/experiments'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import { yamlToExperiment } from 'lib/formikhelpers'

interface LoadFromProps {
  loadCallback?: () => void
}

const LoadFrom: React.FC<LoadFromProps> = ({ loadCallback }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [experiments, setExperiments] = useState<Experiment[]>()
  const [archives, setArchives] = useState<Archive[]>()
  const [predefined, setPredefined] = useState<PreDefinedValue[]>()
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

  const fetchPredefined = async () => setPredefined(await (await getDB()).getAll('predefined'))

  useEffect(() => {
    Promise.all([fetchExperiments(), fetchArchives()])

    fetchPredefined()
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

    if (type === 'p') {
      const experiment = predefined?.filter((p) => p.name === uuid)[0].yaml

      fillExperiment(experiment)

      loadCallback && loadCallback()
      setRadio('')

      return
    }

    const apiRequest = type === 'e' ? api.experiments : api.archives

    setRadio(e.target.value)

    apiRequest
      .detail(uuid)
      .then(({ data }) => {
        fillExperiment(data.kube_object)

        loadCallback && loadCallback()
        setRadio('')

        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'confirm.loadSuccessfully' }),
          })
        )
      })
      .catch(console.error)
  }

  return (
    <Paper>
      <RadioGroup value={radio} onChange={onRadioChange}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
          <Typography>{T('experiments.title')}</Typography>
          <YAML callback={fillExperiment} />
        </Box>
        {experiments && experiments.length > 0 ? (
          <Box display="flex" flexWrap="wrap">
            {experiments.map((e) => (
              <FormControlLabel
                key={e.uid}
                value={`e+${e.uid}`}
                control={<Radio color="primary" />}
                label={RadioLabel(e.name, e.uid)}
              />
            ))}
          </Box>
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
          <Box display="flex" flexWrap="wrap">
            {archives.map((a) => (
              <FormControlLabel
                key={a.uid}
                value={`a+${a.uid}`}
                control={<Radio color="primary" />}
                label={RadioLabel(a.name, a.uid)}
              />
            ))}
          </Box>
        ) : archives?.length === 0 ? (
          <Typography variant="body2">{T('archives.noArchivesFound')}</Typography>
        ) : (
          <SkeletonN n={3} />
        )}
        <Box my={6}>
          <Divider />
        </Box>
        <Box mb={3}>
          <Typography>{T('dashboard.predefined')}</Typography>
        </Box>
        {predefined && predefined.length > 0 ? (
          <Box display="flex" flexWrap="wrap">
            {predefined.map((p) => (
              <FormControlLabel
                key={p.name}
                value={`p+${p.name}`}
                control={<Radio color="primary" />}
                label={RadioLabel(p.name)}
              />
            ))}
          </Box>
        ) : predefined?.length === 0 ? (
          <Typography variant="body2">{T('dashboard.noPredefinedFound')}</Typography>
        ) : (
          <SkeletonN n={3} />
        )}
      </RadioGroup>
    </Paper>
  )
}

export default LoadFrom
