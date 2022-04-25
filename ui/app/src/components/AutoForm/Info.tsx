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

import type { FormikErrors, FormikTouched } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'

import { Belong } from '.'
import { MenuItem } from '@mui/material'
import MoreOptions from 'components/MoreOptions'
import { T } from 'components/T'
import { getIn } from 'formik'
import { useStoreSelector } from 'store'

interface InfoProps {
  belong: Belong
  errors: FormikErrors<Record<string, any>>
  touched: FormikTouched<Record<string, any>>
}

export default function Info({ belong, errors, touched }: InfoProps) {
  const { namespaces } = useStoreSelector((state) => state.experiments)

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
            {namespaces.length && (
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
              getIn(errors, 'name') && getIn(touched, 'name') ? getIn(errors, 'name') : <T id="newE.basic.nameHelper" />
            }
            error={getIn(errors, 'name') && getIn(touched, 'name')}
          />
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
        </>
      )}
    </>
  )
}
