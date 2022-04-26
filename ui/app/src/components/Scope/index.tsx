/*
 * Copyright 2022 Chaos Mesh Authors.
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

import { AutocompleteField, SelectField } from 'components/FormField'
import { MenuItem, Typography } from '@mui/material'
import { arrToObjBySep, objToArrBySep } from 'lib/utils'
import {
  getAnnotations,
  getCommonPodsByNamespaces as getCommonPods,
  getLabels,
  getNetworkTargetPodsByNamespaces as getNetworkTargetPods,
} from 'slices/experiments'
import { getIn, useFormikContext } from 'formik'
import { useEffect, useMemo } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import Mode from './Mode'
import MoreOptions from 'components/MoreOptions'
import ScopePodsTable from './ScopePodsTable'
import type { SelectChangeEvent } from '@mui/material'
import Space from '@ui/mui-extends/esm/Space'
import { T } from 'components/T'

interface ScopeProps {
  namespaces: string[]
  scope?: string
  modeScope?: string
  podsPreviewTitle?: string | JSX.Element
  podsPreviewDesc?: string | JSX.Element
}

const phases = ['Pending', 'Running', 'Succeeded', 'Failed', 'Unknown']

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

  const handleChangeIncludeAll = (e: SelectChangeEvent<string[]>) => {
    const lastValues = getIn(values, e.target.name)
    const currentValues = e.target.value as string[]

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
      <AutocompleteField
        multiple
        name={`${scope}.namespaces`}
        label={<T id="k8s.namespaceSelectors" />}
        helperText={
          getIn(touched, `${scope}.namespaces`) && getIn(errors, `${scope}.namespaces`) ? (
            getIn(errors, `${scope}.namespaces`)
          ) : (
            <T id="newE.scope.namespaceSelectorsHelper" />
          )
        }
        options={!enableKubeSystemNS ? namespaces.filter((d) => d !== 'kube-system') : namespaces}
        error={getIn(errors, `${scope}.namespaces`) && getIn(touched, `${scope}.namespaces`) ? true : false}
        disabled={disabled}
      />

      <AutocompleteField
        multiple
        name={`${scope}.labelSelectors`}
        label={<T id="k8s.labelSelectors" />}
        helperText={<T id="newE.scope.labelSelectorsHelper" />}
        options={labelKVs}
        disabled={disabled}
      />

      <MoreOptions disabled={disabled}>
        <AutocompleteField
          multiple
          name={`${scope}.annotationSelectors`}
          label={<T id="k8s.annotationSelectors" />}
          helperText={<T id="newE.scope.annotationSelectorsHelper" />}
          options={annotationKVs}
          disabled={disabled}
        />

        <SelectField<string[]>
          multiple
          name={`${scope}.podPhaseSelectors`}
          label={<T id="k8s.podPhaseSelectors" />}
          helperText={<T id="newE.scope.phaseSelectorsHelper" />}
          onChange={handleChangeIncludeAll}
          disabled={disabled}
          fullWidth
        >
          <MenuItem value="all">All</MenuItem>
          {phases.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </SelectField>
      </MoreOptions>

      <Mode disabled={disabled} modeScope={modeScope} scope={scope} />

      <div>
        <Typography variant="h6" fontWeight="bold" sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {podsPreviewTitle || <T id="newE.scope.targetPodsPreview" />}
        </Typography>
        <Typography variant="body2" sx={{ color: disabled ? 'text.disabled' : 'text.secondary' }}>
          {podsPreviewDesc || <T id="newE.scope.targetPodsPreviewHelper" />}
        </Typography>
      </div>
      {pods.length > 0 ? (
        <ScopePodsTable scope={scope} pods={pods} />
      ) : (
        <Typography variant="body2" fontWeight="medium" sx={{ color: disabled ? 'text.disabled' : undefined }}>
          <T id="newE.scope.noPodsFound" />
        </Typography>
      )}
    </Space>
  )
}

export default Scope
