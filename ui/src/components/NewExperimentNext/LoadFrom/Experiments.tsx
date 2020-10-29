import { FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { setAction, setBasic, setKind, setTarget } from 'slices/experiments'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import { Experiment } from 'api/experiments.type'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import Wrapper from './Wrapper'
import _snakecase from 'lodash.snakecase'
import api from 'api'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import { yamlToExperiment } from 'lib/formikhelpers'

const Experiments = () => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [experiments, setExperiments] = useState<Experiment[]>()
  const [radio, setRadio] = useState('')

  const fetchExperiments = () =>
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.log)

  useEffect(() => {
    fetchExperiments()
  }, [])

  const onRadioChange = (e: any) => {
    const uuid = e.target.value

    setRadio(uuid)

    api.experiments
      .detail(uuid)
      .then(({ data }) => {
        const y = yamlToExperiment(data.yaml)

        const kind = y.target.kind
        dispatch(setKind(kind))
        dispatch(setAction(y.target[_snakecase(kind)].action ?? ''))
        dispatch(setTarget(y.target))
        dispatch(setBasic(y.basic))
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.importSuccessfully' }),
          })
        )
        dispatch(setAlertOpen(true))
      })
      .catch(console.log)
  }

  return (
    <Wrapper from="experiments">
      <RadioGroup value={radio} onChange={onRadioChange}>
        {experiments && experiments.length > 0 ? (
          experiments.map((e) => (
            <FormControlLabel key={e.uid} value={e.uid} control={<Radio color="primary" />} label={RadioLabel(e)} />
          ))
        ) : experiments?.length === 0 ? (
          <Typography variant="body2">{T('experiments.noExperimentsFound')}</Typography>
        ) : (
          <SkeletonN n={3} />
        )}
      </RadioGroup>
    </Wrapper>
  )
}

export default Experiments
