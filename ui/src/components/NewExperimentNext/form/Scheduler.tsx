import { Box, FormControlLabel, Link, Switch, Typography } from '@material-ui/core'
import { FormikErrors, FormikTouched, getIn } from 'formik'
import React, { useEffect, useState } from 'react'

import { FormattedMessage } from 'react-intl'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import { RootState } from 'store'
import T from 'components/T'
import { TextField } from 'components/FormField'
import Tooltip from 'components-mui/Tooltip'
import { useSelector } from 'react-redux'

const mustBeScheduled = ['pod-kill', 'container-kill']

function validateCron(value: string) {
  let error

  if (value === '') {
    error = 'The cron is required'
  }

  return error
}

function validateDuration(value: string) {
  let error

  if (value === '') {
    error = 'The duration is required'
  }

  return error
}

interface SchedulerProps {
  errors: FormikErrors<Record<string, any>>
  touched: FormikTouched<Record<string, any>>
}

const Scheduler: React.FC<SchedulerProps> = ({ errors, touched }) => {
  const target = useSelector((state: RootState) => state.experiments.target)
  const scheduled = target.kind
    ? target.kind === 'PodChaos' && mustBeScheduled.includes(target.pod_chaos.action)
      ? true
      : false
    : false
  const [immediate, setImmediate] = useState(scheduled)

  useEffect(() => {
    if (scheduled) {
      setImmediate(false)
    }
  }, [scheduled])

  const handleChecked = (_: React.ChangeEvent<HTMLInputElement>, checked: boolean) => {
    if (scheduled) {
      setImmediate(false)
    } else {
      setImmediate(checked)
    }
  }

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography style={{ display: 'flex', alignItems: 'center' }}>
          {T('newE.steps.schedule')}
          <Tooltip
            title={
              <Typography variant="body2">
                <FormattedMessage
                  id="newE.schedule.tooltip"
                  values={{
                    cronv3: (
                      <Link
                        href="https://pkg.go.dev/github.com/robfig/cron/v3"
                        target="_blank"
                        style={{ color: 'white' }}
                        underline="always"
                      >
                        https://pkg.go.dev/github.com/robfig/cron/v3
                      </Link>
                    ),
                  }}
                />
              </Typography>
            }
            arrow
            interactive
          >
            <HelpOutlineIcon fontSize="small" />
          </Tooltip>
        </Typography>
        <Box>
          <FormControlLabel
            style={{ marginLeft: -4, marginRight: 0 }}
            control={
              <Switch name="immediate" color="primary" size="small" checked={immediate} onChange={handleChecked} />
            }
            label={T('newE.schedule.immediate')}
          />
          {scheduled && (
            <Typography variant="subtitle2" color="textSecondary">
              {T('newE.schedule.mustBeScheduled')}
            </Typography>
          )}
        </Box>
      </Box>

      {!immediate && (
        <Box>
          <TextField
            id="scheduler.cron"
            name="scheduler.cron"
            label="Cron"
            validate={validateCron}
            helperText={
              getIn(errors, 'scheduler.cron') && getIn(touched, 'scheduler.cron') ? (
                getIn(errors, 'scheduler.cron')
              ) : (
                <FormattedMessage
                  id="newE.schedule.cronHelper"
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
            error={getIn(errors, 'scheduler.cron') && getIn(touched, 'scheduler.cron') ? true : false}
          />

          {!scheduled && (
            <TextField
              id="scheduler.duration"
              name="scheduler.duration"
              label={T('newE.schedule.duration')}
              validate={validateDuration}
              helperText={
                getIn(errors, 'scheduler.duration') && getIn(touched, 'scheduler.duration')
                  ? getIn(errors, 'scheduler.duration')
                  : T('newE.schedule.durationHelper')
              }
              error={getIn(errors, 'scheduler.duration') && getIn(touched, 'scheduler.duration') ? true : false}
            />
          )}
        </Box>
      )}
    </>
  )
}

export default Scheduler
