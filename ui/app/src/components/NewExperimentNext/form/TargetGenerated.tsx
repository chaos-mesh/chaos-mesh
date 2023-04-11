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
import { MenuItem } from '@mui/material'
import { Form, Formik, FormikErrors, FormikTouched, getIn, setIn } from 'formik'
import { useGetCommonChaosAvailableNamespaces } from 'openapi'
import { useEffect, useState } from 'react'
import { ObjectSchema } from 'yup'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import { Env } from 'slices/experiments'

import { AutocompleteField, LabelField, SelectField, Submit, TextField } from 'components/FormField'
import MoreOptions from 'components/MoreOptions'
import Scope from 'components/Scope'
import i18n from 'components/T'

import basicData from '../data/basic'
import { Kind, Spec } from '../data/types'
import { Stale } from 'api/queryUtils'

interface TargetGeneratedProps {
  env: Env
  kind?: Kind | ''
  data: Spec
  validationSchema?: ObjectSchema
  onSubmit: (values: Record<string, any>) => void
}

const TargetGenerated: React.FC<TargetGeneratedProps> = ({ env, kind, data, validationSchema, onSubmit }) => {
  const { spec } = useStoreSelector((state) => state.experiments)

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  let initialValues = Object.entries(data).reduce((acc, [k, v]) => {
    if (v instanceof Object && v.field) {
      acc = setIn(acc, k, v.value)
    } else {
      acc[k] = v
    }

    return acc
  }, {} as Record<string, any>)

  if (env === 'k8s' && kind === 'NetworkChaos') {
    const action = initialValues.action
    delete initialValues.action
    const direction = initialValues.direction
    delete initialValues.direction
    const externalTargets = initialValues.externalTargets
    delete initialValues.externalTargets

    initialValues = {
      action,
      [action]: action !== 'partition' ? initialValues : undefined,
      direction,
      externalTargets,
    }
  }

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    setInit({
      ...initialValues,
      ...spec,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [spec])

  const parseDataToFormFields = (
    errors: FormikErrors<Record<string, any>>,
    touched: FormikTouched<Record<string, any>>
  ) => {
    const rendered = Object.entries(data)
      .filter(([_, v]) => v && v instanceof Object && v.field)
      .map(([k, v]) => {
        if (env === 'k8s' && kind === 'NetworkChaos' && k !== 'direction' && k !== 'externalTargets') {
          k = `${data.action}.${k}`
        }

        if (env === 'k8s' && kind === 'BlockChaos' && k !== 'volumeName' && k !== 'iops' && k !== 'containerNames') {
          k = `${data.action}.${k}`
        }

        switch (v.field) {
          case 'text':
            return (
              <TextField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(touched, k) && getIn(errors, k)}
                {...v.inputProps}
              />
            )
          case 'textarea':
            return (
              <TextField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(touched, k) && getIn(errors, k)}
                multiline
                rows={6}
                {...v.inputProps}
              />
            )
          case 'number':
            return (
              <TextField
                key={k}
                type="number"
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(errors, k) && getIn(touched, k)}
                {...v.inputProps}
              />
            )
          case 'select':
            return (
              <SelectField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(errors, k) && getIn(touched, k)}
              >
                {v.items?.map((option: string | { label: string; value: any }) =>
                  option instanceof Object ? (
                    <MenuItem key={option.label} value={option.value}>
                      {option.label}
                    </MenuItem>
                  ) : (
                    <MenuItem key={option} value={option}>
                      {option}
                    </MenuItem>
                  )
                )}
              </SelectField>
            )
          case 'label':
            return (
              <LabelField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(errors, k) && getIn(touched, k)}
              />
            )
          case 'autocomplete':
            return (
              <AutocompleteField
                multiple
                key={k}
                name={k}
                label={v.label}
                helperText={v.helperText}
                options={v.items!}
              />
            )
          default:
            return null
        }
      })
      .filter((d) => d)

    return <>{rendered.map((d) => d)}</>
  }

  return (
    <Formik enableReinitialize initialValues={init} validationSchema={validationSchema} onSubmit={onSubmit}>
      {({ values, setFieldValue, errors, touched }) => {
        const beforeTargetOpen = () => {
          if (!getIn(values, 'target')) {
            setFieldValue('target', {
              selector: basicData.spec.selector,
              mode: basicData.spec.mode,
              value: basicData.spec.value,
            })
          }
        }

        const afterTargetClose = () => {
          if (getIn(values, 'target')) {
            setFieldValue('target', undefined)
          }
        }

        return (
          <Form>
            <Space>{parseDataToFormFields(errors, touched)}</Space>
            {env === 'k8s' && kind === 'NetworkChaos' && (
              <MoreOptions
                title={i18n('newE.target.network.target.title')}
                beforeOpen={beforeTargetOpen}
                afterClose={afterTargetClose}
              >
                {values.target && (
                  <Scope
                    env="k8s"
                    kind={kind}
                    namespaces={namespaces!}
                    scope="target.selector"
                    modeScope="target"
                    previewTitle={i18n('newE.target.network.target.podsPreview')}
                  />
                )}
              </MoreOptions>
            )}
            <Submit />
          </Form>
        )
      }}
    </Formik>
  )
}

export default TargetGenerated
