import React from 'react'
import { Box, MenuItem, FormLabel, FormControlLabel, Radio, RadioGroup } from '@material-ui/core'
import { SelectField, TextField } from '../../../components/FormField'

import { StepProps } from '../types'

interface ScopeProps extends StepProps {
  namespaces: string[]
}
// TODO: fake data
const phases = ['all', 'running', 'pending', 'failed']
const modes = ['all', 'fixed number', 'fixed percentage', 'max percentage']

function upperFirst(str: string) {
  if (!str) return ''

  return str.charAt(0).toUpperCase() + str.slice(1)
}

export default function ScopeStep({ formProps, namespaces }: ScopeProps) {
  const { values, handleChange, handleBlur } = formProps

  return (
    <Box maxWidth="30rem" mx="auto">
      <SelectField
        id="scope.namespaceSelector"
        name="scope.namespaceSelector"
        label="Namespace Selector"
        labelId="scope.namespaceSelector-label"
        helperText="Multiple"
        multiple
        value={values.scope.namespaceSelector}
        onChange={handleChange}
      >
        {namespaces.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      {/* TODO: Label Selector: {[key: string]: string} */}

      {/* TODO: Annotation Selector: {[key: string]: string} */}

      {/* TODO: Field Selector: {[key: string]: string} */}

      <SelectField
        id="scope.phaseSelector"
        name="scope.phaseSelector"
        label="Phase Selector"
        labelId="scope.phaseSelector-label"
        helperText="Multiple"
        multiple
        value={values.scope.phaseSelector}
        onChange={handleChange}
      >
        {phases.map((option: string) => (
          <MenuItem key={option} value={option}>
            {upperFirst(option)}
          </MenuItem>
        ))}
      </SelectField>

      <Box mb={4}>
        <FormLabel component="label">Mode</FormLabel>
        <Box display="flex" justifyContent="space-between">
          <RadioGroup
            id="scope.mode"
            name="scope.mode"
            aria-label="mode"
            style={{ flexBasis: '60%' }}
            value={values.scope.mode}
            onChange={handleChange}
          >
            {modes.map((m: string) => {
              return <FormControlLabel key={m} value={m} control={<Radio />} label={upperFirst(m)} />
            })}
          </RadioGroup>

          {values.scope.mode !== 'all' && (
            <TextField
              id="scope.value"
              label="Value"
              type="number"
              autoComplete="off"
              helperText="Please input a value"
              value={values.scope.value}
              onBlur={handleBlur}
              onChange={handleChange}
            />
          )}
        </Box>
      </Box>
    </Box>
  )
}
