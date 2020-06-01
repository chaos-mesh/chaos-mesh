import { InputAdornment, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import React from 'react'
import { StepperFormProps } from '../types'
import { upperFirst } from 'lib/utils'

interface ScopeStepProps {
  formProps: StepperFormProps
  namespaces: string[]
}

const phases = ['all', 'pending', 'running', 'succeeded', 'failed', 'unknown']
const modes = ['all', { name: 'Random one', value: 'one' }, 'fixed number', 'fixed percent', 'random max percent']
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

const ScopeStep: React.FC<ScopeStepProps> = ({ formProps, namespaces }) => {
  const { values, handleChange } = formProps

  const handleChangeIncludeAll = (id: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    const lastValues = id.split('.').reduce((acc, cur) => acc[cur], values as any)
    const currentValues = (e.target.value as unknown) as string[]

    if (!lastValues.includes('all') && currentValues.includes('all')) {
      e.target.value = ['all'] as any
    }

    if (lastValues.includes('all') && currentValues.length > 1) {
      e.target.value = currentValues.filter((v) => v !== 'all') as any
    }

    handleChange(e)
  }

  return (
    <>
      <SelectField
        id="scope.namespace_selectors"
        name="scope.namespace_selectors"
        label="Namespace Selectors"
        helperText="Multiple options"
        multiple
        value={values.scope.namespace_selectors}
        onChange={handleChange}
      >
        {namespaces.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      <TextField
        id="scope.label_selectors"
        label="Label selectors"
        helperText="Json of label selectors"
        multiline
        value={values.scope.label_selectors}
        onChange={handleChange}
      />

      {/* TODO: Annotation Selectors: {[key: string]: string} */}

      {/* TODO: Field Selectors: {[key: string]: string} */}

      <SelectField
        id="scope.mode"
        name="scope.mode"
        label="Mode"
        helperText="Select a mode"
        value={values.scope.mode}
        onChange={handleChange}
      >
        {modes.map((option) =>
          typeof option === 'string' ? (
            <MenuItem key={option} value={option.split(' ').join('-')}>
              {upperFirst(option)}
            </MenuItem>
          ) : (
            <MenuItem key={option.value} value={option.value}>
              {option.name}
            </MenuItem>
          )
        )}
      </SelectField>

      {values.scope.mode !== 'all' && values.scope.mode !== 'one' && (
        <TextField
          id="scope.value"
          label="Mode Value"
          helperText="Please fill the mode value"
          value={values.scope.value}
          onChange={handleChange}
          InputProps={{
            endAdornment: modesWithAdornment.includes(values.scope.mode) && (
              <InputAdornment position="end">%</InputAdornment>
            ),
          }}
        />
      )}

      <AdvancedOptions>
        <SelectField
          id="scope.phase_selectors"
          name="scope.phase_selectors"
          label="Phase Selectors"
          helperText="Multiple options"
          multiple
          value={values.scope.phase_selectors}
          onChange={handleChangeIncludeAll('scope.phase_selectors')}
        >
          {phases.map((option: string) => (
            <MenuItem key={option} value={option}>
              {upperFirst(option)}
            </MenuItem>
          ))}
        </SelectField>
      </AdvancedOptions>
    </>
  )
}

export default ScopeStep
