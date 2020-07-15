import { SelectField, TextField } from 'components/FormField'

import { MenuItem } from '@material-ui/core'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { toTitleCase } from 'lib/utils'

const actions = ['pod failure', 'pod kill', 'container kill']

export default function Pod(props: StepperFormTargetProps) {
  const { values, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.pod_chaos.action"
        name="target.pod_chaos.action"
        label="Action"
        helperText="Select a PodChaos action"
        value={values.target.pod_chaos.action}
        onChange={handleActionChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option.split(' ').join('-')}>
            {toTitleCase(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.pod_chaos.action === 'container-kill' && (
        <TextField
          id="target.pod_chaos.container_name"
          name="target.pod_chaos.container_name"
          label="Container Name"
          helperText="Fill the container name you want to kill"
        />
      )}
    </>
  )
}
