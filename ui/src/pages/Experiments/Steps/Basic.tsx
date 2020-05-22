import React from 'react'
import { Box, MenuItem } from '@material-ui/core'
import { TextField, SelectField } from '../../../components/FormField'

import { StepProps } from '../types'

interface BasicProps extends StepProps {
  namespaces: string[]
}

export default function BasicStep({ formProps, namespaces }: BasicProps) {
  const { values, handleBlur, handleChange } = formProps

  return (
    <Box maxWidth="30rem" mx="auto">
      <TextField
        id="basic.name"
        label="Name"
        type="text"
        autoComplete="off"
        helperText="Please input an experiment name"
        value={values.basic.name}
        onBlur={handleBlur}
        onChange={handleChange}
      />

      <SelectField
        id="basic.namespace"
        name="basic.namespace"
        label="Namespace"
        labelId="basic.namespace-label"
        helperText="Please select an experiment namespace"
        value={values.basic.namespace}
        onChange={handleChange}
      >
        {namespaces.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      {/* TODO: Labels: {[key: string]: string} */}
    </Box>
  )
}
