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
import { MenuItem } from '@mui/material'
import { useFormikContext } from 'formik'
import { getIn } from 'formik'
import { useGetCommonChaosAvailableNamespaces } from 'openapi'

import { LabelField, SelectField, TextField } from 'components/FormField'
import MoreOptions from 'components/MoreOptions'
import { T } from 'components/T'

import { Belong } from '.'
import { isInstant } from './validation'
import { Stale } from 'api/queryUtils'

interface InfoProps {
  belong: Belong
  kind: string
  action?: string
}

export default function Info({ belong, kind, action }: InfoProps) {
  const { errors, touched } = useFormikContext()

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  return (
    <>
      {(belong === Belong.Experiment || belong === Belong.Schedule) && (
        <>
          <TextField
            fast
            name="metadata.name"
            label={<T id="common.name" />}
            helperText={
              getIn(errors, 'metadata.name') && getIn(touched, 'metadata.name') ? (
                getIn(errors, 'metadata.name')
              ) : (
                <T id={`${belong === Belong.Schedule ? 'newS' : 'newE'}.basic.nameHelper`} />
              )
            }
            error={getIn(errors, 'metadata.name') && getIn(touched, 'metadata.name')}
          />
          <MoreOptions>
            {namespaces && (
              <SelectField
                name="metadata.namespace"
                label={<T id="k8s.namespace" />}
                helperText={<T id="newE.basic.namespaceHelper" />}
              >
                {namespaces.map((n) => (
                  <MenuItem key={n} value={n}>
                    {n}
                  </MenuItem>
                ))}
              </SelectField>
            )}
            <LabelField name="metadata.labels" label={<T id="k8s.labels" />} />
            <LabelField name="metadata.annotations" label={<T id="k8s.annotations" />} />
          </MoreOptions>
        </>
      )}
      {belong === Belong.Workflow && (
        <>
          <TextField
            fast
            name="name"
            label={<T id="common.name" />}
            helperText={
              getIn(errors, 'name') && getIn(touched, 'name') ? getIn(errors, 'name') : <T id="newW.node.nameHelper" />
            }
            error={getIn(errors, 'name') && getIn(touched, 'name')}
          />
          {!isInstant(kind, action) && (
            <TextField
              fast
              name="deadline"
              label={<T id="newW.node.deadline" />}
              helperText={
                getIn(errors, 'deadline') && getIn(touched, 'deadline') ? (
                  getIn(errors, 'deadline')
                ) : (
                  <T id="newW.node.deadlineHelper" />
                )
              }
              error={getIn(errors, 'deadline') && getIn(touched, 'deadline')}
            />
          )}
        </>
      )}
    </>
  )
}
