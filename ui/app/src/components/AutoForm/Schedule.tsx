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
import { InputAdornment, Link, MenuItem } from '@mui/material'
import { useFormikContext } from 'formik'
import { getIn } from 'formik'

import { SelectField, TextField } from 'components/FormField'
import { T } from 'components/T'

export default function Schedule() {
  const { errors, touched } = useFormikContext()

  return (
    <>
      <TextField
        fast
        name="schedule"
        label="schedule"
        helperText={
          getIn(errors, 'schedule') && getIn(touched, 'schedule') ? (
            getIn(errors, 'schedule')
          ) : (
            <T
              id="newS.basic.scheduleHelper"
              values={{
                crontabguru: (
                  <Link href="https://crontab.guru/" target="_blank" underline="always">
                    https://crontab.guru/
                  </Link>
                ),
              }}
            />
          )
        }
        error={getIn(errors, 'spec.schedule') && getIn(touched, 'spec.schedule') ? true : false}
      />
      <TextField
        fast
        type="number"
        name="historyLimit"
        label="historyLimit"
        helperText={
          getIn(errors, 'historyLimit') && getIn(touched, 'historyLimit') ? (
            getIn(errors, 'historyLimit')
          ) : (
            <T id="newS.basic.historyLimitHelper" />
          )
        }
        error={getIn(errors, 'historyLimit') && getIn(touched, 'historyLimit')}
      />
      <SelectField
        name="concurrencyPolicy"
        label="concurrencyPolicy"
        helperText={
          getIn(errors, 'concurrencyPolicy') && getIn(touched, 'concurrencyPolicy') ? (
            getIn(errors, 'concurrencyPolicy')
          ) : (
            <T id="newS.basic.concurrencyPolicyHelper" />
          )
        }
        error={getIn(errors, 'concurrencyPolicy') && getIn(touched, 'concurrencyPolicy')}
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
        name="startingDeadlineSeconds"
        label="startingDeadlineSeconds"
        endAdornment={
          <InputAdornment position="end">
            <T id="common.seconds" />
          </InputAdornment>
        }
        helperText={
          getIn(errors, 'startingDeadlineSeconds') && getIn(touched, 'startingDeadlineSeconds') ? (
            getIn(errors, 'startingDeadlineSeconds')
          ) : (
            <T id="newS.basic.startingDeadlineSecondsHelper" />
          )
        }
        error={getIn(errors, 'startingDeadlineSeconds') && getIn(touched, 'startingDeadlineSeconds')}
      />
    </>
  )
}
