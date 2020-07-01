import { Box, Divider, FormControlLabel, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, makeStyles } from '@material-ui/core/styles'

import { Experiment } from '../types'
import { TextField } from 'components/FormField'
import { mustSchedule } from 'lib/formikhelpers'
import { useFormikContext } from 'formik'

const useStyles = makeStyles((theme: Theme) => ({
  scheduleTitle: {
    margin: `${theme.spacing(3)} 0`,
  },
  pre: {
    background: theme.palette.grey[200],
    overflowX: 'auto',
    '& code': {
      whiteSpace: 'pre',
    },
  },
}))

const ScheduleStep: React.FC = () => {
  const classes = useStyles()

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
        <Typography className={classes.scheduleTitle}>Schedule</Typography>
        <pre className={classes.pre}>
          <code>
            {`
  Field name   | Mandatory? | Allowed values  | Allowed special characters
  ----------   | ---------- | --------------  | --------------------------
  Seconds      | Yes        | 0-59            | * / , -
  Minutes      | Yes        | 0-59            | * / , -
  Hours        | Yes        | 0-23            | * / , -
  Day of month | Yes        | 1-31            | * / , - ?
  Month        | Yes        | 1-12 or JAN-DEC | * / , -
  Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?
          `}
          </code>
        </pre>
        <pre className={classes.pre}>
          <code>
            {`
  Entry                  | Description                                | Equivalent To
  -----                  | -----------                                | -------------
  @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
  @monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
  @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
  @daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
  @hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
          `}
          </code>
        </pre>

        <TextField
          id="scheduler.cron"
          name="scheduler.cron"
          label="Cron"
          helperText="You can use https://crontab.guru/ to help generate your cron syntax and confirm what time it will run"
        />

        <TextField
          id="scheduler.duration"
          name="scheduler.duration"
          label="Duration"
          helperText="The Experiment duration"
        />
      </Box>
    </>
  )
}

export default ScheduleStep
