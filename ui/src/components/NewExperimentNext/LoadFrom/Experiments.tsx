import { FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Experiment } from 'api/experiments.type'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import Wrapper from './Wrapper'
import api from 'api'
import { yamlToExperiment } from 'lib/formikhelpers'

const Experiments = () => {
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

    // api.experiments
    //   .detail(uuid)
    //   .then(({ data }) => setInitialValues(yamlToExperiment(data.yaml)))
    //   .catch(console.log)
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
