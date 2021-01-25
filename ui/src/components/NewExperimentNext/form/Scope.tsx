import { AutocompleteMultipleField, SelectField, TextField } from 'components/FormField'
import { Box, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import React, { useEffect, useMemo, useRef } from 'react'
import { arrToObjBySep, joinObjKVs, toTitleCase } from 'lib/utils'
import {
  getAnnotations,
  getCommonPodsByNamespaces as getCommonPods,
  getLabels,
  getNetworkTargetPodsByNamespaces as getNetworkTargetPods,
} from 'slices/experiments'
import { getIn, useFormikContext } from 'formik'
import { useStoreDispatch, useStoreSelector } from 'store'

import AdvancedOptions from 'components/AdvancedOptions'
import ScopePodsTable from './ScopePodsTable'
import T from 'components/T'

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

const ScopeStep: React.FC<ScopeStepProps> = ({ namespaces, scope = 'scope', podsPreviewTitle, podsPreviewDesc }) => {
  const { values, handleChange, setFieldValue, errors, touched } = useFormikContext()
  const {
    namespace_selectors: currentNamespaces,
    label_selectors: currentLabels,
    annotation_selectors: currentAnnotations,
  } = getIn(values, scope)

  const state = useStoreSelector((state) => state)
  const { enableKubeSystemNS } = state.settings
  const { labels, annotations } = state.experiments
  const pods = scope === 'scope' ? state.experiments.pods : state.experiments.networkTargetPods
  const getPods = scope === 'scope' ? getCommonPods : getNetworkTargetPods
  const dispatch = useStoreDispatch()

  const kvSeparator = ': '
  const labelKVs = useMemo(() => joinObjKVs(labels, kvSeparator), [labels])
  const annotationKVs = useMemo(() => joinObjKVs(annotations, kvSeparator), [annotations])

  const firstRender = useRef(true)

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
    // Set ns selectors directly when CLUSTER_MODE=false.
    if (namespaces.length === 1) {
      setFieldValue(`${scope}.namespace_selectors`, namespaces)

      if (scope === 'scope') {
        setFieldValue('namespace', namespaces[0])
      }
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
  }, [currentNamespaces, getPods, dispatch])

  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false

      return
    }

    dispatch(
      getPods({
        namespace_selectors: currentNamespaces,
        label_selectors: arrToObjBySep(currentLabels, kvSeparator),
        annotation_selectors: arrToObjBySep(currentAnnotations, kvSeparator),
      })
    )

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [firstRender, currentLabels, currentAnnotations])

  return (
    <>
      <AutocompleteMultipleField
        id={`${scope}.namespace_selectors`}
        name={`${scope}.namespace_selectors`}
        label={T('newE.scope.namespaceSelectors')}
        helperText={
          getIn(touched, `${scope}.namespace_selectors`) && getIn(errors, `${scope}.namespace_selectors`)
            ? getIn(errors, `${scope}.namespace_selectors`)
            : T('common.multiOptions')
        }
        options={!enableKubeSystemNS ? namespaces.filter((d) => d !== 'kube-system') : namespaces}
        error={
          getIn(errors, `${scope}.namespace_selectors`) && getIn(touched, `${scope}.namespace_selectors`) ? true : false
        }
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
