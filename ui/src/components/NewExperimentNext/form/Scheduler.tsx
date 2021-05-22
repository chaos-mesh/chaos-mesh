import { Box, FormControlLabel, Switch, Typography } from '@material-ui/core'
import { FormikErrors, FormikTouched, getIn } from 'formik'
import React, { useEffect, useState } from 'react'

import { RootState } from 'store'
import T from 'components/T'
import { TextField } from 'components/FormField'
import { useSelector } from 'react-redux'
import { validateDuration } from 'lib/formikhelpers'

const mustBeScheduled = ['pod-kill', 'container-kill']

interface SchedulerProps {
  errors: FormikErrors<Record<string, any>>
  touched: FormikTouched<Record<string, any>>
}

const Scheduler: React.FC<SchedulerProps> = ({ errors, touched }) => {
  const { fromExternal, basic, target } = useSelector((state: RootState) => state.experiments)
  const scheduled = target.kind
    ? target.kind === 'PodChaos' && mustBeScheduled.includes(target.pod_chaos.action)
      ? true
      : false
    : false
  const [continuous, setContinuous] = useState(scheduled)

  useEffect(() => {
    if (scheduled) {
      setContinuous(false)
    }
  }, [scheduled])

  useEffect(() => {
    if (fromExternal && basic.scheduler.cron === '' && basic.scheduler.duration === '') {
      setContinuous(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fromExternal])

  const handleChecked = (_: React.ChangeEvent<HTMLInputElement>, checked: boolean) => {
    if (scheduled) {
      setContinuous(false)
    } else {
      setContinuous(checked)
    }
  }

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography>{T('newE.steps.schedule')}</Typography>
        <Box>
          <FormControlLabel
            style={{ marginRight: 0 }}
            control={
              <Switch name="continuous" color="primary" size="small" checked={continuous} onChange={handleChecked} />
            }
            label={T('newE.schedule.continuous')}
          />
          {scheduled && (
            <Typography variant="subtitle2" color="textSecondary">
              {T('newE.schedule.mustBeScheduled')}
            </Typography>
          )}
        </Box>
      </Box>

      {!continuous && !scheduled && (
        <TextField
          fast
          name="scheduler.duration"
          label={T('newE.schedule.duration')}
          validate={validateDuration()}
          helperText={
            getIn(errors, 'scheduler.duration') && getIn(touched, 'scheduler.duration')
              ? getIn(errors, 'scheduler.duration')
              : T('newE.schedule.durationHelper')
          }
          error={getIn(errors, 'scheduler.duration') && getIn(touched, 'scheduler.duration') ? true : false}
        />
      )}
    </>
  )
}

export default Scheduler
