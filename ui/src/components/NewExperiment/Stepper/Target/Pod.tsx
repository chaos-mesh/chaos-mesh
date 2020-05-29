import { Box, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

import React from 'react'
import { StepperFormProps } from 'components/NewExperiment/types'

// TODO: fake data, maybe use object to map option description
const actions = ['Killing Pod', 'Pod unavailable in a specified period of time', 'Killing Container']

export default function PodPanel(props: StepperFormProps) {
  const { values, handleBlur, handleChange } = props

  return (
    <Box maxWidth="30rem" mx="auto">
      <SelectField
        id="target.pod.action"
        name="target.pod.action"
        label="Action"
        labelId="target.pod.action-label"
        helperText="Please select the action to attack"
        value={values.target.pod.action}
        onChange={handleChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.pod.action === 'Killing Container' && (
        <TextField
          id="target.pod.container"
          label="Container Name"
          type="text"
          autoComplete="off"
          helperText="Please input a container name"
          value={values.target.pod.container}
          onBlur={handleBlur}
          onChange={handleChange}
        />
      )}
    </Box>
  )
}
