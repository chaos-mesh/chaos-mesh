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

import { AutocompleteMultipleField, SelectField } from 'components/FormField'
import { Divider, MenuItem, Typography } from '@mui/material'
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
import OtherOptions from 'components/OtherOptions'
import ScopePodsTable from './ScopePodsTable'
import Space from '@ui/mui-extends/esm/Space'
import i18n from 'components/T'

interface ScopeProps {
  namespaces: string[]
  scope?: string
  modeScope?: string
  podsPreviewTitle?: string | JSX.Element
  podsPreviewDesc?: string | JSX.Element
}

const phases = [{ label: 'All', value: 'all' }, 'Pending', 'Running', 'Succeeded', 'Failed', 'Unknown']

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
        label={i18n('k8s.namespaceSelectors')}
        helperText={
          getIn(touched, `${scope}.namespaces`) && getIn(errors, `${scope}.namespaces`)
            ? getIn(errors, `${scope}.namespaces`)
            : i18n('common.multiOptions')
        }
        options={!enableKubeSystemNS ? namespaces.filter((d) => d !== 'kube-system') : namespaces}
        error={getIn(errors, `${scope}.namespaces`) && getIn(touched, `${scope}.namespaces`) ? true : false}
        disabled={disabled}
      />

      <AutocompleteMultipleField
        name={`${scope}.labelSelectors`}
        label={i18n('k8s.labelSelectors')}
        helperText={i18n('common.multiOptions')}
        options={labelKVs}
        disabled={disabled}
      />

      <OtherOptions disabled={disabled}>
        <AutocompleteMultipleField
          name={`${scope}.annotationSelectors`}
          label={i18n('k8s.annotationsSelectors')}
          helperText={i18n('common.multiOptions')}
          options={annotationKVs}
          disabled={disabled}
        />

        <SelectField
          name={`${scope}.podPhaseSelectors`}
          label={i18n('k8s.podPhaseSelectors')}
          helperText={i18n('common.multiOptions')}
          multiple
          onChange={handleChangeIncludeAll}
          disabled={disabled}
        >
          {phases.map((option) =>
            typeof option === 'string' ? (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ) : (
              <MenuItem key={option.value} value={option.value}>
                {option.label}
              </MenuItem>
            )
          )}
        </SelectField>
      </OtherOptions>

      <Divider />
      <Typography>{i18n('newE.scope.mode')}</Typography>
      <Mode disabled={disabled} modeScope={modeScope} scope={scope} />
      <Divider />

      <div>
        <Typography sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {podsPreviewTitle || i18n('newE.scope.targetPodsPreview')}
        </Typography>
        <Typography variant="body2" sx={{ color: disabled ? 'text.disabled' : 'text.secondary' }}>
          {podsPreviewDesc || i18n('newE.scope.targetPodsPreviewHelper')}
        </Typography>
      </div>
      {pods.length > 0 ? (
        <ScopePodsTable scope={scope} pods={pods} />
      ) : (
        <Typography variant="subtitle2" sx={{ color: disabled ? 'text.disabled' : undefined }}>
          {i18n('newE.scope.noPodsFound')}
        </Typography>
      )}
    </Space>
  )
}

export default Scope
