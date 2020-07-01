import { InputAdornment, MenuItem } from '@material-ui/core'
import React, { useMemo } from 'react'
import { SelectField, TextField } from 'components/FormField'
import { joinObjKVs, upperFirst } from 'lib/utils'

import AdvancedOptions from 'components/AdvancedOptions'
import { Experiment } from '../types'
import { useFormikContext } from 'formik'

interface ScopeStepProps {
  namespaces: string[]
  labels: { [key: string]: string[] }
  annotations: { [key: string]: string[] }
}

const phases = ['all', 'pending', 'running', 'succeeded', 'failed', 'unknown']
const modes = ['all', { name: 'Random one', value: 'one' }, 'fixed number', 'fixed percent', 'random max percent']
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

const ScopeStep: React.FC<ScopeStepProps> = ({ namespaces, labels, annotations }) => {
  const { values, handleChange } = useFormikContext<Experiment>()

  const labelKVs = useMemo(() => joinObjKVs(labels, ': '), [labels])
  const annotationKVs = useMemo(() => joinObjKVs(annotations, ': '), [annotations])

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
      >
        {namespaces.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      <SelectField
        id="scope.label_selectors"
        name="scope.label_selectors"
        label="Label Selectors"
        helperText="Multiple options"
        multiple
      >
        {labelKVs.map((option: string) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </SelectField>

      <SelectField id="scope.mode" name="scope.mode" label="Mode" helperText="Select a mode">
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
          name="scope.value"
          label="Mode Value"
          helperText="Please fill the mode value"
          InputProps={{
            endAdornment: modesWithAdornment.includes(values.scope.mode) && (
              <InputAdornment position="end">%</InputAdornment>
            ),
          }}
        />
      )}

      <AdvancedOptions>
        <SelectField
          id="scope.annotation_selectors"
          name="scope.annotation_selectors"
          label="Annotation Selectors"
          helperText="Multiple options"
          multiple
        >
          {annotationKVs.map((option: string) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </SelectField>

        <SelectField
          id="scope.phase_selectors"
          name="scope.phase_selectors"
          label="Phase Selectors"
          helperText="Multiple options"
          multiple
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
