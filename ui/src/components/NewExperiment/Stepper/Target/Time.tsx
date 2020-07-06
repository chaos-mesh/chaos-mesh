import { LabelField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { resetOtherChaos } from 'lib/formikhelpers'

export default function Kernel(props: StepperFormTargetProps) {
  useEffect(() => {
    resetOtherChaos(props, 'TimeChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <LabelField
        id="target.time_chaos.clock_ids"
        name="target.time_chaos.clock_ids"
        label="Clock ids"
        helperText="Type string and end with a space to generate the clock ids"
      />
      <TextField id="target.time_chaos.offset" name="target.time_chaos.offset" label="Offset" />

      <AdvancedOptions>
        <LabelField
          id="target.time_chaos.container_names"
          name="target.time_chaos.container_names"
          label="Container names"
          helperText="Type string and end with a space to generate the container names"
        />
      </AdvancedOptions>
    </>
  )
}
