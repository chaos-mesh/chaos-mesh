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

import {
  Env,
  clearPods,
  getAnnotations,
  getCommonPods,
  getLabels,
  getNetworkTargetPods,
  getPhysicalMachines,
} from 'slices/experiments'

import { podPhases } from 'components/AutoForm/data'
import { AutocompleteField, SelectField } from 'components/FormField'
import MoreOptions from 'components/MoreOptions'
import { T } from 'components/T'

import { arrToObjBySep, objToArrBySep } from 'lib/utils'

import DeprecatedAddress from './DeprecatedAddress'
import Mode from './Mode'
import TargetsTable from './TargetsTable'

interface ScopeProps {
  env: Env
  namespaces: string[]
  scope?: string
  modeScope?: string
  previewTitle?: string | JSX.Element
}

const Scope = ({ env, namespaces, scope = 'selector', modeScope = '', previewTitle }: ScopeProps) => {
  const { values, setFieldValue, errors, touched } = useFormikContext()
  const {
    namespaces: currentNamespaces,
    labelSelectors: currentLabels,
    annotationSelectors: currentAnnotations,
  } = getIn(values, scope)

  const { settings, experiments } = useStoreSelector((state) => state)
  const { enableKubeSystemNS } = settings
  const { labels, annotations, pods: storePods, networkTargetPods, physicalMachines } = experiments
  const isTargetField = scope.startsWith('target')
  const pods = !isTargetField ? storePods : networkTargetPods
  const getPods = !isTargetField ? getCommonPods : getNetworkTargetPods
  const targets = env === 'k8s' ? pods : physicalMachines
  const getTargets = env === 'k8s' ? getPods : getPhysicalMachines // Get different targets according to the env.
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
    // Set namespaces automatically when `targetNamespace` is set because there is only one namespace.
    if (namespaces.length === 1) {
      setFieldValue(`${scope}.namespace`, namespaces)

      // Set namespace in metadata automatically too.
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
        getTargets({
          namespaces: currentNamespaces,
          labelSelectors: arrToObjBySep(currentLabels, kvSeparator),
          annotationSelectors: arrToObjBySep(currentAnnotations, kvSeparator),
        })
      )
    }
  }, [dispatch, getTargets, currentNamespaces, currentLabels, currentAnnotations])

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
        options={enableKubeSystemNS ? namespaces : namespaces.filter((d) => d !== 'kube-system')}
        error={getIn(errors, `${scope}.namespaces`) && getIn(touched, `${scope}.namespaces`) ? true : false}
      />

      <AutocompleteField
        freeSolo
        multiple
        name={`${scope}.labelSelectors`}
        label={<T id="k8s.labelSelectors" />}
        helperText={<T id="newE.scope.labelSelectorsHelper" />}
        options={labelKVs}
      />

      <MoreOptions>
        <AutocompleteField
          freeSolo
          multiple
          name={`${scope}.annotationSelectors`}
          label={<T id="k8s.annotationSelectors" />}
          helperText={<T id="newE.scope.annotationSelectorsHelper" />}
          options={annotationKVs}
        />

        <SelectField<string[]>
          multiple
          name={`${scope}.podPhaseSelectors`}
          label={<T id="k8s.podPhaseSelectors" />}
          helperText={<T id="newE.scope.phaseSelectorsHelper" />}
          fullWidth
        >
          {podPhases.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </SelectField>
      </MoreOptions>

      <Mode modeScope={modeScope} scope={scope} />

      <div>
        <Typography fontWeight="medium">
          {previewTitle || <T id={`newE.scope.target${env === 'k8s' ? 'Pods' : 'PhysicalMachines'}Preview`} />}
        </Typography>
        <Typography variant="body2" sx={{ color: 'text.secondary' }}>
          <T id={`newE.scope.target${env === 'k8s' ? 'Pods' : 'PhysicalMachines'}PreviewHelper`} />
        </Typography>
      </div>

      {targets.length > 0 ? (
        <TargetsTable env={env} scope={scope} data={targets} />
      ) : (
        <Typography variant="body2" fontWeight="medium">
          <T id={`newE.scope.no${env === 'k8s' ? 'Pods' : 'PhysicalMachines'}Found`} />
        </Typography>
      )}
    </Space>
  )
}

interface ConditionalScopeProps extends ScopeProps {
  kind: string
}

const ConditionalScope = ({ kind, ...rest }: ConditionalScopeProps) => {
  const disabled = kind === 'AWSChaos' || kind === 'GCPChaos'

  const { useNewPhysicalMachine } = useStoreSelector((state) => state.settings)

  if (disabled) {
    return (
      <Typography
        variant="body2"
        sx={{ color: 'text.disabled' }}
      >{`${kind} does not need to define the scope.`}</Typography>
    )
  }

  if (rest.env === 'physic' && !useNewPhysicalMachine) {
    return <DeprecatedAddress />
  }

  return <Scope {...rest} />
}

export default ConditionalScope
