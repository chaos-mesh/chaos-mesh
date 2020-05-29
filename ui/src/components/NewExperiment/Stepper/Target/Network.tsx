import React from 'react'

import { Box, MenuItem } from '@material-ui/core'

import { SelectField, TextField } from 'components/FormField'
import { StepperFormProps } from 'components/NewExperiment/types'

// TODO: fake data, maybe use object to map option description
const actions = ['Network delay', 'Losing tcp packets', '...']

export default function NetworkPanel(props: StepperFormProps) {
  const { values, handleBlur, handleChange } = props

  return (
    <Box maxWidth="30rem" mx="auto">
      <SelectField
        id="target.network.action"
        name="target.network.action"
        label="Action"
        labelId="target.network.action-label"
        helperText="Please select the action to attack"
        value={values.target.network.action}
        onChange={handleChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>
      {values.target.network.action === 'Network delay' && (
        <>
          <TextField
            id="target.network.delay.latency"
            label="Latency"
            type="text"
            autoComplete="off"
            helperText="Please input the latency"
            value={values.target.network.delay.latency}
            onBlur={handleBlur}
            onChange={handleChange}
          />

          <TextField
            id="target.network.delay.correlation"
            label="Correlation"
            type="text"
            autoComplete="off"
            helperText="Please input the correlation"
            value={values.target.network.delay.correlation}
            onBlur={handleBlur}
            onChange={handleChange}
          />
          <TextField
            id="target.network.delay.jitter"
            label="Jitter"
            type="text"
            autoComplete="off"
            helperText="Please input the jitter"
            value={values.target.network.delay.jitter}
            onBlur={handleBlur}
            onChange={handleChange}
          />
        </>
      )}
    </Box>
  )
}
