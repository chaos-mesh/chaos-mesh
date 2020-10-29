import { FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { setAction, setBasic, setKind, setTarget } from 'slices/experiments'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import { Archive } from 'api/archives.type'
import RadioLabel from './RadioLabel'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import Wrapper from './Wrapper'
import _snakecase from 'lodash.snakecase'
import api from 'api'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import { yamlToExperiment } from 'lib/formikhelpers'

const Archives = () => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

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
    <Wrapper from="archives">
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
