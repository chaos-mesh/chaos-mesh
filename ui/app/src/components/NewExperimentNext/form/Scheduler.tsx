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
import { Box, FormControlLabel, Link, Switch, Typography } from '@mui/material'
import { FormikErrors, FormikTouched, getIn, useFormikContext } from 'formik'
import { useEffect, useState } from 'react'
import { FormattedMessage } from 'react-intl'

import { useStoreSelector } from 'store'

import { TextField } from 'components/FormField'
import { ExperimentKind } from 'components/NewExperiment/types'
import i18n from 'components/T'

import { validateDuration, validateSchedule } from 'lib/formikhelpers'

function isInstant(kind: ExperimentKind | '', action: string) {
  if (kind === 'PodChaos' && (action === 'pod-kill' || action === 'container-kill')) {
    return true
  }

  return false
}

interface SchedulerProps {
  errors: FormikErrors<Record<string, any>>
  touched: FormikTouched<Record<string, any>>
  inSchedule?: boolean
}

const Scheduler: React.FC<SchedulerProps> = ({ errors, touched, inSchedule = false }) => {
  const { fromExternal, kindAction, basic } = useStoreSelector((state) => state.experiments)
  const { values, setFieldValue } = useFormikContext()
  const [kind, action] = kindAction
  const instant = isInstant(kind, action)

  const [continuous, setContinuous] = useState(false)

  useEffect(() => {
    if (!inSchedule && fromExternal && basic.spec.duration === '') {
      setContinuous(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fromExternal])

  const handleChecked = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked

    setContinuous(checked)

    if (checked && getIn(values, 'spec.duration') !== '') {
      setFieldValue('spec.duration', '')
    }
  }

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography fontWeight={500}>{i18n('newE.steps.run')}</Typography>
        {!inSchedule && (
          <FormControlLabel
            style={{ marginRight: 0 }}
            control={
              <Switch name="continuous" color="primary" size="small" checked={continuous} onChange={handleChecked} />
            }
            label={i18n('newE.run.continuous')}
            disabled={instant}
          />
        )}
      </Box>

      {inSchedule && (
        <TextField
          fast
          name="spec.schedule"
          label={i18n('schedules.single')}
          validate={validateSchedule()}
          helperText={
            getIn(errors, 'spec.schedule') && getIn(touched, 'spec.schedule') ? (
              getIn(errors, 'spec.schedule')
            ) : (
              <FormattedMessage
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
      )}

      {!continuous && (
        <TextField
          name="spec.duration"
          label={i18n('common.duration')}
          validate={instant ? undefined : validateDuration()}
          helperText={
            getIn(errors, 'spec.duration') && getIn(touched, 'spec.duration')
              ? getIn(errors, 'spec.duration')
              : i18n('common.durationHelper')
          }
          error={getIn(errors, 'spec.duration') && getIn(touched, 'spec.duration') ? true : false}
          disabled={instant}
        />
      )}
    </>
  )
}

export default Scheduler
