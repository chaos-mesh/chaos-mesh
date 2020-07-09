import { LabelField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { Typography } from '@material-ui/core'
import { resetOtherChaos } from 'lib/formikhelpers'

export default function Stress(props: StepperFormTargetProps) {
  useEffect(() => {
    resetOtherChaos(props, 'StressChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <Typography gutterBottom>CPU</Typography>
      <TextField
        type="number"
        id="target.stress_chaos.stressors.cpu.workers"
        name="target.stress_chaos.stressors.cpu.workers"
        label="Workers"
        helperText="CPU workers"
      />
      <TextField
        type="number"
        id="target.stress_chaos.stressors.cpu.load"
        name="target.stress_chaos.stressors.cpu.load"
        label="Load"
        helperText="CPU load"
      />
      <LabelField
        id="target.stress_chaos.stressors.cpu.options"
        name="target.stress_chaos.stressors.cpu.options"
        label="Options of CPU stressors"
        helperText="Type string and end with a space to generate the stress-ng options"
      />

      <Typography gutterBottom>Memory</Typography>
      <TextField
        type="number"
        id="target.stress_chaos.stressors.memory.workers"
        name="target.stress_chaos.stressors.memory.workers"
        label="Workers"
        helperText="Memory workers"
      />
      <TextField
        id="target.stress_chaos.stressors.memory.size"
        name="target.stress_chaos.stressors.memory.size"
        label="Size"
        helperText="Memory size"
      />
      <LabelField
        id="target.stress_chaos.stressors.memory.options"
        name="target.stress_chaos.stressors.memory.options"
        label="Options of Memory stressors"
        helperText="Type string and end with a space to generate the stress-ng options"
      />

      <AdvancedOptions>
        <TextField
          id="target.stress_chaos.stressng_stressors"
          name="target.stress_chaos.stressng_stressors"
          label="Options of stress-ng"
          helperText="The options of stress-ng, treated as a string"
        />
      </AdvancedOptions>
    </>
  )
}
