import { LabelField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import { FormikCtx } from 'components/NewExperiment/types'
import T from 'components/T'
import { resetOtherChaos } from 'lib/formikhelpers'
import { useFormikContext } from 'formik'

export default function Time() {
  const formikCtx: FormikCtx = useFormikContext()
  const { errors, touched } = formikCtx

  useEffect(() => {
    resetOtherChaos(formikCtx, 'TimeChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <TextField
        id="target.time_chaos.time_offset"
        name="target.time_chaos.time_offset"
        label={T('newE.target.time.offset')}
        helperText={T('newE.target.time.offsetHelper')}
        error={errors.target?.time_chaos?.time_offset && touched.target?.time_chaos?.time_offset ? true : false}
      />

      <AdvancedOptions>
        <LabelField
          id="target.time_chaos.clock_ids"
          name="target.time_chaos.clock_ids"
          label="Clock ids"
          helperText="Optional. Type string and end with a space to generate the clock ids. If it's empty, it will be set to ['CLOCK_REALTIME']"
        />
        <LabelField
          id="target.time_chaos.container_names"
          name="target.time_chaos.container_names"
          label="Affected container names"
          helperText="Optional. Type string and end with a space to generate the container names. If it's empty, all containers will be injected"
        />
      </AdvancedOptions>
    </>
  )
}
