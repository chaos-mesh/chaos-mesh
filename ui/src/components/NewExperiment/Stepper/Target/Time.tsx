import { LabelField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { resetOtherChaos } from 'lib/formikhelpers'

export default function Time(props: StepperFormTargetProps) {
  useEffect(() => {
    resetOtherChaos(props, 'TimeChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <TextField
        id="target.time_chaos.offset"
        name="target.time_chaos.offset"
        label="Offset"
        helperText="The time offset"
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
