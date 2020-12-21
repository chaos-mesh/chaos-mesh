import { AutocompleteMultipleField, SelectField, TextField } from 'components/FormField'
import { Box, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import React, { useEffect, useMemo } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { arrToObjBySep, joinObjKVs, toTitleCase } from 'lib/utils'
import { getAnnotations, getLabels, getPodsByNamespaces as getPods } from 'slices/experiments'
import { getIn, useFormikContext } from 'formik'

import AdvancedOptions from 'components/AdvancedOptions'
import ScopePodsTable from './ScopePodsTable'
import T from 'components/T'
import { useSelector } from 'react-redux'

interface ScopeStepProps {
  namespaces: string[]
  scope?: string
  podsPreviewTitle?: string | JSX.Element
  podsPreviewDesc?: string | JSX.Element
}

const phases = ['all', 'pending', 'running', 'succeeded', 'failed', 'unknown']
const modes = [
  { name: 'Random One', value: 'one' },
  { name: 'Fixed Number', value: 'fixed' },
  { name: 'Fixed Percent', value: 'fixed-percent' },
  { name: 'Random Max Percent', value: 'random-max-percent' },
]
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

const labelFilters = ['pod-template-hash']

const ScopeStep: React.FC<ScopeStepProps> = ({ namespaces, scope = 'scope', podsPreviewTitle, podsPreviewDesc }) => {
  const { values, handleChange, setFieldValue } = useFormikContext()
  const {
    namespace_selectors: currentNamespaces,
    label_selectors: currentLabels,
    annotation_selectors: currentAnnotations,
  } = getIn(values, scope)

  const { labels, annotations, pods } = useSelector((state: RootState) => state.experiments)
  const dispatch = useStoreDispatch()

  const kvSeparator = ': '
  const labelKVs = useMemo(() => joinObjKVs(labels, kvSeparator, labelFilters), [labels])
  const annotationKVs = useMemo(() => joinObjKVs(annotations, kvSeparator), [annotations])

  const handleChangeIncludeAll = (id: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    const lastValues = getIn(values, id)
    const currentValues = (e.target.value as unknown) as string[]

    if (!lastValues.includes('all') && currentValues.includes('all')) {
      e.target.value = ['all'] as any
    }

    if (lastValues.includes('all') && currentValues.length > 1) {
      e.target.value = currentValues.filter((v) => v !== 'all') as any
    }

    handleChange(e)
  }

  useEffect(() => {
    if (namespaces.length === 1) {
      setFieldValue(`${scope}.namespace_selectors`, namespaces)

      if (scope === 'scope') {
        setFieldValue('namespace', namespaces[0])
      }
    } else {
      setFieldValue(`${scope}.namespace_selectors`, ['default'])
    }
  }, [namespaces, scope, setFieldValue])

  useEffect(() => {
    if (currentNamespaces.length) {
      dispatch(
        getPods({
          namespace_selectors: currentNamespaces,
        })
      )

      dispatch(getLabels(currentNamespaces))
      dispatch(getAnnotations(currentNamespaces))
    }
  }, [currentNamespaces, dispatch])

  useEffect(() => {
    dispatch(
      getPods({
        namespace_selectors: currentNamespaces,
        label_selectors: arrToObjBySep(currentLabels, kvSeparator),
        annotation_selectors: arrToObjBySep(currentAnnotations, kvSeparator),
      })
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentLabels, currentAnnotations])

  return (
    <>
      <AutocompleteMultipleField
        id={`${scope}.namespace_selectors`}
        name={`${scope}.namespace_selectors`}
        label={T('newE.scope.namespaceSelectors')}
        helperText={T('common.multiOptions')}
        options={namespaces}
      />

      <AutocompleteMultipleField
        id={`${scope}.label_selectors`}
        name={`${scope}.label_selectors`}
        label={T('k8s.labelSelectors')}
        helperText={T('common.multiOptions')}
        options={labelKVs}
      />

      <AdvancedOptions>
        <AutocompleteMultipleField
          id={`${scope}.annotation_selectors`}
          name={`${scope}.annotation_selectors`}
          label={T('k8s.annotationsSelectors')}
          helperText={T('common.multiOptions')}
          options={annotationKVs}
        />

        <SelectField
          id={`${scope}.mode`}
          name={`${scope}.mode`}
          label={T('newE.scope.mode')}
          helperText={T('newE.scope.modeHelper')}
        >
          <MenuItem value="all">All</MenuItem>
          {modes.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.name}
            </MenuItem>
          ))}
        </SelectField>

        {getIn(values, scope).mode !== 'all' && getIn(values, scope).mode !== 'one' && (
          <TextField
            id={`${scope}.value`}
            name={`${scope}.value`}
            label={T('newE.scope.modeValue')}
            helperText={T('newE.scope.modeValueHelper')}
            InputProps={{
              endAdornment: modesWithAdornment.includes(getIn(values, scope).mode) && (
                <InputAdornment position="end">%</InputAdornment>
              ),
            }}
          />
        )}

        <SelectField
          id={`${scope}.phase_selectors`}
          name={`${scope}.phase_selectors`}
          label={T('k8s.phaseSelectors')}
          helperText={T('common.multiOptions')}
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

      <Box mb={3}>
        <Typography>{podsPreviewTitle || T('newE.scope.affectedPodsPreview')}</Typography>
        <Typography variant="subtitle2" color="textSecondary">
          {podsPreviewDesc || T('newE.scope.affectedPodsPreviewHelper')}
        </Typography>
      </Box>

      {pods.length > 0 ? (
        <ScopePodsTable scope={scope} pods={pods} />
      ) : (
        <Typography variant="subtitle2">{T('newE.scope.noPodsFound')}</Typography>
      )}
    </>
  )
}

export default ScopeStep
