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
import { MenuItem, Typography } from '@mui/material'
import { getIn, useFormikContext } from 'formik'
import { useEffect, useMemo } from 'react'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { clearPods, getAnnotations, getCommonPods, getLabels, getNetworkTargetPods } from 'slices/experiments'

import { podPhases } from 'components/AutoForm/data'
import { AutocompleteField, SelectField } from 'components/FormField'
import MoreOptions from 'components/MoreOptions'
import { T } from 'components/T'

import { arrToObjBySep, objToArrBySep } from 'lib/utils'

import Mode from './Mode'
import ScopePodsTable from './ScopePodsTable'

interface ScopeProps {
  kind?: string
  namespaces: string[]
  scope?: string
  modeScope?: string
  podsPreviewTitle?: string | JSX.Element
}

const Scope: React.FC<ScopeProps> = ({ kind, namespaces, scope = 'selector', modeScope = '', podsPreviewTitle }) => {
  const { values, setFieldValue, errors, touched } = useFormikContext()
  const {
    namespaces: currentNamespaces,
    labelSelectors: currentLabels,
    annotationSelectors: currentAnnotations,
  } = getIn(values, scope)

  const state = useStoreSelector((state) => state)
  const { enableKubeSystemNS } = state.settings
  const { labels, annotations } = state.experiments
  const isTargetField = scope.startsWith('target')
  const pods = !isTargetField ? state.experiments.pods : state.experiments.networkTargetPods
  const getPods = !isTargetField ? getCommonPods : getNetworkTargetPods
  const disabled = kind === 'AWSChaos' || kind === 'GCPChaos'
  const dispatch = useStoreDispatch()

  const kvSeparator = ': '
  const labelKVs = useMemo(() => objToArrBySep(labels, kvSeparator), [labels])
  const annotationKVs = useMemo(() => objToArrBySep(annotations, kvSeparator), [annotations])

  useEffect(() => {
    return () => {
      dispatch(clearPods())
    }
  }, [dispatch])

  useEffect(() => {
    // Set ns selectors automatically when `CLUSTER_MODE=false` because there is only one namespace.
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
  }, [dispatch, currentNamespaces])

  useEffect(() => {
    if (currentNamespaces.length) {
      dispatch(
        getPods({
          namespaces: currentNamespaces,
          labelSelectors: arrToObjBySep(currentLabels, kvSeparator),
          annotationSelectors: arrToObjBySep(currentAnnotations, kvSeparator),
        })
      )
    }
  }, [dispatch, getPods, currentNamespaces, currentLabels, currentAnnotations])

  return (
    <Space>
      <AutocompleteField
        freeSolo
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
        freeSolo
        multiple
        name={`${scope}.labelSelectors`}
        label={<T id="k8s.labelSelectors" />}
        helperText={<T id="newE.scope.labelSelectorsHelper" />}
        options={labelKVs}
        disabled={disabled}
      />

      <MoreOptions disabled={disabled}>
        <AutocompleteField
          freeSolo
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
          disabled={disabled}
          fullWidth
        >
          {podPhases.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </SelectField>
      </MoreOptions>

      <Mode disabled={disabled} modeScope={modeScope} scope={scope} />

      <div>
        <Typography fontWeight="bold" sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {podsPreviewTitle || <T id="newE.scope.targetPodsPreview" />}
        </Typography>
        <Typography variant="body2" sx={{ color: disabled ? 'text.disabled' : 'text.secondary' }}>
          <T id="newE.scope.targetPodsPreviewHelper" />
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
