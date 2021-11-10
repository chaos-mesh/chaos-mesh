/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { AutocompleteMultipleField, SelectField, TextField } from 'components/FormField'
import { InputAdornment, MenuItem, Typography } from '@material-ui/core'
import { arrToObjBySep, objToArrBySep, toTitleCase } from 'lib/utils'
import {
  getAnnotations,
  getCommonPodsByNamespaces as getCommonPods,
  getLabels,
  getNetworkTargetPodsByNamespaces as getNetworkTargetPods,
} from 'slices/experiments'
import { getIn, useFormikContext } from 'formik'
import { useEffect, useMemo } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import OtherOptions from 'components/OtherOptions'
import ScopePodsTable from './ScopePodsTable'
import Space from 'components-mui/Space'
import T from 'components/T'

interface ScopeProps {
  namespaces: string[]
  scope?: string
  modeScope?: string
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

const Scope: React.FC<ScopeProps> = ({
  namespaces,
  scope = 'spec.selector',
  modeScope = 'spec',
  podsPreviewTitle,
  podsPreviewDesc,
}) => {
  const { values, handleChange, setFieldValue, errors, touched } = useFormikContext()
  const {
    namespaces: currentNamespaces,
    labelSelectors: currentLabels,
    annotationSelectors: currentAnnotations,
  } = getIn(values, scope)

  const state = useStoreSelector((state) => state)
  const { enableKubeSystemNS } = state.settings
  const { labels, annotations, kindAction } = state.experiments
  const [kind] = kindAction
  const pods = scope === 'spec.selector' ? state.experiments.pods : state.experiments.networkTargetPods
  const getPods = scope === 'spec.selector' ? getCommonPods : getNetworkTargetPods
  const disabled = kind === 'AWSChaos' || kind === 'GCPChaos'
  const dispatch = useStoreDispatch()

  const kvSeparator = ': '
  const labelKVs = useMemo(() => objToArrBySep(labels, kvSeparator), [labels])
  const annotationKVs = useMemo(() => objToArrBySep(annotations, kvSeparator), [annotations])

  const handleChangeIncludeAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    const lastValues = getIn(values, e.target.name)
    const currentValues = e.target.value as unknown as string[]

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
      setFieldValue(`${scope}.namespace`, namespaces)

      if (scope === 'spec.selector') {
        setFieldValue('namespace', namespaces[0])
      }
    }
  }, [namespaces, scope, setFieldValue])

  useEffect(() => {
    if (currentNamespaces.length) {
      dispatch(getLabels(currentNamespaces))
      dispatch(getAnnotations(currentNamespaces))
    }
  }, [dispatch, getPods, currentNamespaces])

  useEffect(() => {
    if (currentNamespaces.length) {
      dispatch(
        getPods({
          namespaces: currentNamespaces,
          labelSelectors: arrToObjBySep(currentLabels, kvSeparator) as any,
          annotationSelectors: arrToObjBySep(currentAnnotations, kvSeparator) as any,
        })
      )
    }
  }, [dispatch, getPods, currentNamespaces, currentLabels, currentAnnotations])

  return (
    <Space>
      <AutocompleteMultipleField
        name={`${scope}.namespaces`}
        label={T('k8s.namespaceSelectors')}
        helperText={
          getIn(touched, `${scope}.namespaces`) && getIn(errors, `${scope}.namespaces`)
            ? getIn(errors, `${scope}.namespaces`)
            : T('common.multiOptions')
        }
        options={!enableKubeSystemNS ? namespaces.filter((d) => d !== 'kube-system') : namespaces}
        error={getIn(errors, `${scope}.namespaces`) && getIn(touched, `${scope}.namespaces`) ? true : false}
        disabled={disabled}
      />

      <AutocompleteMultipleField
        name={`${scope}.labelSelectors`}
        label={T('k8s.labelSelectors')}
        helperText={T('common.multiOptions')}
        options={labelKVs}
        disabled={disabled}
      />

      <OtherOptions disabled={disabled}>
        <AutocompleteMultipleField
          name={`${scope}.annotationSelectors`}
          label={T('k8s.annotationsSelectors')}
          helperText={T('common.multiOptions')}
          options={annotationKVs}
          disabled={disabled}
        />

        <SelectField
          name={`${modeScope}.mode`}
          label={T('newE.scope.mode')}
          helperText={T('newE.scope.modeHelper')}
          disabled={disabled}
        >
          <MenuItem value="all">All</MenuItem>
          {modes.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.name}
            </MenuItem>
          ))}
        </SelectField>

        {!['all', 'one'].includes(getIn(values, modeScope).mode) && (
          <TextField
            name={`${modeScope}.value`}
            label={T('newE.scope.modeValue')}
            helperText={T('newE.scope.modeValueHelper')}
            InputProps={{
              endAdornment: modesWithAdornment.includes(getIn(values, scope).mode) && (
                <InputAdornment position="end">%</InputAdornment>
              ),
            }}
            disabled={disabled}
          />
        )}

        <SelectField
          name={`${scope}.podPhaseSelectors`}
          label={T('k8s.podPhaseSelectors')}
          helperText={T('common.multiOptions')}
          multiple
          onChange={handleChangeIncludeAll}
          disabled={disabled}
        >
          {phases.map((option: string) => (
            <MenuItem key={option} value={option}>
              {toTitleCase(option)}
            </MenuItem>
          ))}
        </SelectField>
      </OtherOptions>

      <div>
        <Typography sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {podsPreviewTitle || T('newE.scope.targetPodsPreview')}
        </Typography>
        <Typography variant="body2" sx={{ color: disabled ? 'text.disabled' : 'text.secondary' }}>
          {podsPreviewDesc || T('newE.scope.targetPodsPreviewHelper')}
        </Typography>
      </div>
      {pods.length > 0 ? (
        <ScopePodsTable scope={scope} pods={pods} />
      ) : (
        <Typography variant="subtitle2" sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {T('newE.scope.noPodsFound')}
        </Typography>
      )}
    </Space>
  )
}

export default Scope
