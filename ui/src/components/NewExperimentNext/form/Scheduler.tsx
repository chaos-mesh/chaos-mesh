import { Box, FormControlLabel, Link, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'

import { FormattedMessage } from 'react-intl'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import { RootState } from 'store'
import T from 'components/T'
import { TextField } from 'components/FormField'
import Tooltip from 'components-mui/Tooltip'
import { useSelector } from 'react-redux'

const mustBeScheduled = ['pod-kill', 'container-kill']

interface SchedulerProps {}

const Scheduler: React.FC<SchedulerProps> = () => {
  const target = useSelector((state: RootState) => state.experiments.target)
  const scheduled = target.kind
    ? target.kind === 'PodChaos' && mustBeScheduled.includes(target.pod_chaos.action)
      ? true
      : false
    : false
  const [immediate, setImmediate] = useState(scheduled)

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
        <Typography>
          {T('newE.steps.schedule')}{' '}
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
            style={{ verticalAlign: 'sub' }}
            arrow
            interactive
          >
            <HelpOutlineIcon fontSize="small" />
          </Tooltip>
        </Typography>
        <Box>
          <FormControlLabel
            style={{ marginRight: 0 }}
            control={<Switch name="immediate" color="primary" checked={immediate} onChange={handleChecked} />}
            label={T('newE.schedule.immediate')}
          />
          {scheduled && (
            <Typography variant="subtitle2" color="textSecondary">
              {T('newE.schedule.mustBeScheduled')}
            </Typography>
          )}
        </Box>
      </Box>

      <Box hidden={immediate}>
        <TextField
          id="scheduler.cron"
          name="scheduler.cron"
          label="Cron"
          helperText={
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
          }
        />

        {!scheduled && (
          <TextField
            id="scheduler.duration"
            name="scheduler.duration"
            label={T('newE.schedule.duration')}
            helperText={T('newE.schedule.durationHelper')}
          />
        )}
      </Box>
    </>
  )
}

export default Scheduler
