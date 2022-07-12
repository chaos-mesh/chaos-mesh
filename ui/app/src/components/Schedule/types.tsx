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
import { InputAdornment, MenuItem } from '@mui/material'
import { FormikProps, FormikValues, getIn } from 'formik'
import { number, string } from 'yup'

import { SelectField, TextField } from 'components/FormField'
import { T } from 'components/T'

export interface ScheduleSpecific {
  schedule: string
  historyLimit?: number
  concurrencyPolicy?: 'Forbid' | 'Allow'
  startingDeadlineSeconds?: number
}

export const data: ScheduleSpecific = {
  schedule: '',
  historyLimit: 1,
  concurrencyPolicy: 'Forbid',
  startingDeadlineSeconds: undefined,
}

export const Fields = ({ errors, touched }: Pick<FormikProps<FormikValues>, 'errors' | 'touched'>) => (
  <>
    <TextField
      fast
      type="number"
      name="spec.historyLimit"
      label={<T id="newS.basic.historyLimit" />}
      helperText={
        getIn(errors, 'spec.historyLimit') && getIn(touched, 'spec.historyLimit') ? (
          getIn(errors, 'spec.historyLimit')
        ) : (
          <T id="newS.basic.historyLimitHelper" />
        )
      }
      error={getIn(errors, 'spec.historyLimit') && getIn(touched, 'spec.historyLimit')}
    />
    <SelectField
      name="spec.concurrencyPolicy"
      label={<T id="newS.basic.concurrencyPolicy" />}
      helperText={
        getIn(errors, 'spec.concurrencyPolicy') && getIn(touched, 'spec.concurrencyPolicy') ? (
          getIn(errors, 'spec.concurrencyPolicy')
        ) : (
          <T id="newS.basic.concurrencyPolicyHelper" />
        )
      }
      error={getIn(errors, 'spec.concurrencyPolicy') && getIn(touched, 'spec.concurrencyPolicy')}
    >
      <MenuItem value="Forbid">
        <T id="newS.basic.forbid" />
      </MenuItem>
      <MenuItem value="Allow">
        <T id="newS.basic.allow" />
      </MenuItem>
    </SelectField>
    <TextField
      fast
      type="number"
      name="spec.startingDeadlineSeconds"
      label={<T id="newS.basic.startingDeadlineSeconds" />}
      endAdornment={
        <InputAdornment position="end">
          <T id="common.seconds" />
        </InputAdornment>
      }
      helperText={
        getIn(errors, 'spec.startingDeadlineSeconds') && getIn(touched, 'spec.startingDeadlineSeconds') ? (
          getIn(errors, 'spec.startingDeadlineSeconds')
        ) : (
          <T id="newS.basic.startingDeadlineSecondsHelper" />
        )
      }
      error={getIn(errors, 'spec.startingDeadlineSeconds') && getIn(touched, 'spec.startingDeadlineSeconds')}
    />
  </>
)

export const schema = {
  historyLimit: number().min(1, 'The historyLimit is at least 1'),
  concurrencyPolicy: string().required('The concurrencyPolicy is required'),
  startingDeadlineSeconds: number().min(0, 'The startingDeadlineSeconds is at least 0').nullable(true),
}
