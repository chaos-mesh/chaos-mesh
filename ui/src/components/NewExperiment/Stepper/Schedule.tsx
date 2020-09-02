import { Box, Divider, FormControlLabel, Link, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'

import { Experiment } from '../types'
import HelpOutlineIcon from '@material-ui/icons/HelpOutline'
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
        label="Immediate"
      />
      {mustBeScheduled && (
        <Typography variant="subtitle2" color="textSecondary">
          The action you chose must be scheduled.
        </Typography>
      )}

      <Box hidden={isImmediate} mt={3}>
        <Divider />
        <Box my={3}>
          <Typography>
            Schedule{' '}
            <Tooltip
              title={
                <Typography variant="body2">
                  Chaos Mesh use{' '}
                  <Link
                    href="https://github.com/robfig/cron/v3"
                    target="_blank"
                    style={{ color: 'white' }}
                    underline="always"
                  >
                    github.com/robfig/cron/v3
                  </Link>{' '}
                  to define the schedule. View its doc for more details.
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
          helperText="You can use https://crontab.guru/ to help generate your cron syntax and confirm what time it will run"
        />

        {values.target.pod_chaos.action !== 'pod-kill' && values.target.pod_chaos.action !== 'container-kill' && (
          <TextField
            id="scheduler.duration"
            name="scheduler.duration"
            label="Duration"
            helperText="The experiment duration"
          />
        )}
      </Box>
    </>
  )
}

export default ScheduleStep
