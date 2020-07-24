import { AutocompleteMultipleField, SelectField, TextField } from 'components/FormField'
import { InputAdornment, MenuItem } from '@material-ui/core'
import React, { useMemo } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { getAnnotations, getLabels } from 'slices/experiments'
import { getIn, useFormikContext } from 'formik'
import { joinObjKVs, toTitleCase } from 'lib/utils'

import AdvancedOptions from 'components/AdvancedOptions'
import { Experiment } from '../types'
import { useSelector } from 'react-redux'

interface ScopeStepProps {
  namespaces: string[]
  scope?: string
}

const phases = ['all', 'pending', 'running', 'succeeded', 'failed', 'unknown']
const modes = ['all', { name: 'Random One', value: 'one' }, 'fixed number', 'fixed percent', 'random max percent']
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

const ScopeStep: React.FC<ScopeStepProps> = ({ namespaces, scope = 'scope' }) => {
  const { values, handleChange } = useFormikContext<Experiment>()

  const { labels, annotations } = useSelector((state: RootState) => state.experiments)
  const storeDispatch = useStoreDispatch()

  const labelKVs = useMemo(() => joinObjKVs(labels, ': '), [labels])
  const annotationKVs = useMemo(() => joinObjKVs(annotations, ': '), [annotations])

  const handleNamespaceSelectorsChangeCallback = (labels: string[]) => {
    const _labels = labels.length !== 0 ? labels : namespaces

    storeDispatch(getLabels(_labels.join(',')))
    storeDispatch(getAnnotations(_labels.join(',')))
  }

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
      <AutocompleteMultipleField
        id={`${scope}.namespace_selectors`}
        name={`${scope}.namespace_selectors`}
        label="Namespace Selectors"
        helperText="Multiple options"
        options={namespaces}
        onChangeCallback={handleNamespaceSelectorsChangeCallback}
      />

      <AutocompleteMultipleField
        id={`${scope}.label_selectors`}
        name={`${scope}.label_selectors`}
        label="Label Selectors"
        helperText="Multiple options"
        options={labelKVs}
      />

      <SelectField id={`${scope}.mode`} name={`${scope}.mode`} label="Mode" helperText="Select the experiment mode">
        {modes.map((option) =>
          typeof option === 'string' ? (
            <MenuItem key={option} value={option.split(' ').join('-')}>
              {toTitleCase(option)}
            </MenuItem>
          ) : (
            <MenuItem key={option.value} value={option.value}>
              {option.name}
            </MenuItem>
          )
        )}
      </SelectField>

      {getIn(values, scope).mode !== 'all' && getIn(values, scope).mode !== 'one' && (
        <TextField
          id={`${scope}.value`}
          name={`${scope}.value`}
          label="Mode Value"
          helperText="Please fill the mode value"
          InputProps={{
            endAdornment: modesWithAdornment.includes(getIn(values, scope).mode) && (
              <InputAdornment position="end">%</InputAdornment>
            ),
          }}
        />
      )}

      <AdvancedOptions>
        <AutocompleteMultipleField
          id={`${scope}.annotation_selectors`}
          name={`${scope}.annotation_selectors`}
          label="Annotation Selectors"
          helperText="Multiple options"
          options={annotationKVs}
        />

        <SelectField
          id={`${scope}.phase_selectors`}
          name={`${scope}.phase_selectors`}
          label="Phase Selectors"
          helperText="Multiple options"
          multiple
          onChange={handleChangeIncludeAll(`${scope}.phase_selectors`)}
        >
          {phases.map((option: string) => (
            <MenuItem key={option} value={option}>
              {toTitleCase(option)}
            </MenuItem>
          ))}
        </SelectField>
      </AdvancedOptions>
    </>
  )
}

export default ScopeStep
