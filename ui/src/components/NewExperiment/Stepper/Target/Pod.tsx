import { SelectField, TextField } from 'components/FormField'

import { MenuItem } from '@material-ui/core'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { upperFirst } from 'lib/utils'

const actions = ['pod kill', 'pod failure', 'container kill']

export default function PodPanel(props: StepperFormTargetProps) {
  const { values, handleChange, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.pod_chaos.action"
        name="target.pod_chaos.action"
        label="Action"
        helperText="Please select an action"
        value={values.target.pod_chaos.action}
        onChange={handleActionChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option.split(' ').join('-')}>
            {upperFirst(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.pod_chaos.action === 'container-kill' && (
        <TextField
          id="target.pod_chaos.container_name"
          label="Container Name"
          helperText="Input the container name you want to kill"
          value={values.target.pod_chaos.container_name}
          onChange={handleChange}
        />
      )}
    </>
  )
}
