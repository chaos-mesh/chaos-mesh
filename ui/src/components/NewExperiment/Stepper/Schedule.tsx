import { Box, Divider, FormControlLabel, Link, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'

import { Experiment } from '../types'
import { FormattedMessage } from 'react-intl'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
import T from 'components/T'
import { TextField } from 'components/FormField'
import Tooltip from 'components/Tooltip'
import { mustSchedule } from 'lib/formikhelpers'
import { useFormikContext } from 'formik'

const ScheduleStep: React.FC = () => {
  const { values } = useFormikContext<Experiment>()
  const hasScheduled = values.scheduler.cron !== '' || values.scheduler.duration !== ''
  const mustBeScheduled = mustSchedule(values)
  const immediate = mustBeScheduled ? false : hasScheduled ? false : true
  const [isImmediate, setIsImmediate] = useState(immediate)

  const handleChecked = (_: React.ChangeEvent<HTMLInputElement>, checked: boolean) => {
    if (mustBeScheduled) {
      setIsImmediate(false)
    } else {
      setIsImmediate(checked)
    }
  }

  return (
    <>
      <FormControlLabel
        control={<Switch name="immediate" color="primary" checked={isImmediate} onChange={handleChecked} />}
        label={T('newE.schedule.immediate')}
      />
      {mustBeScheduled && (
        <Typography variant="subtitle2" color="textSecondary">
          {T('newE.schedule.mustBeScheduled')}
        </Typography>
      )}

      <Box hidden={isImmediate} mt={3}>
        <Divider />
        <Box my={3}>
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
        </Box>

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

        {values.target.pod_chaos.action !== 'pod-kill' && values.target.pod_chaos.action !== 'container-kill' && (
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

export default ScheduleStep
