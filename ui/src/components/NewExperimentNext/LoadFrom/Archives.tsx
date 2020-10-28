import { FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import Wrapper from './Wrapper'
import api from 'api'
import { yamlToExperiment } from 'lib/formikhelpers'

const Archives = () => {
  const [archives, setArchives] = useState<Archive[]>()
  const [radio, setRadio] = useState('')

  const fetchArchives = () =>
    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.log)

  useEffect(() => {
    fetchArchives()
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
        {archives && archives.length > 0 ? (
          archives.map((a) => (
            <FormControlLabel key={a.uid} value={a.uid} control={<Radio color="primary" />} label={RadioLabel(a)} />
          ))
        ) : archives?.length === 0 ? (
          <Typography variant="body2">{T('archives.no_archives_found')}</Typography>
        ) : (
          <SkeletonN n={3} />
        )}
      </RadioGroup>
    </Wrapper>
  )
}

export default Archives
