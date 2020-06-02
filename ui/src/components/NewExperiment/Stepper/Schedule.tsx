import { Box, Divider, FormControlLabel, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, makeStyles } from '@material-ui/core/styles'

import { StepperFormProps } from '../types'
import { TextField } from 'components/FormField'

const useStyles = makeStyles((theme: Theme) => ({
  cronTitle: {
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

interface ScheduleStepProps {
  formProps: StepperFormProps
}

const ScheduleStep: React.FC<ScheduleStepProps> = ({ formProps }) => {
  const classes = useStyles()

  const [isImmediate, setIsImmediate] = useState(true)
  const { values, handleChange } = formProps

  const handleChecked = (_: React.ChangeEvent<HTMLInputElement>, checked: boolean) => setIsImmediate(checked)

  return (
    <>
      <FormControlLabel
        control={<Switch name="immediate" color="primary" checked={isImmediate} onChange={handleChecked} />}
        label="Immediate Job"
      />

      <Box hidden={isImmediate} mt={3}>
        <Divider />
        <Typography className={classes.cronTitle}>Cron Job</Typography>
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
          label="Cron"
          helperText="Schedule crontab"
          value={values.scheduler.cron}
          onChange={handleChange}
        />

        <TextField
          id="scheduler.duration"
          label="Duration"
          helperText="Schedule duration"
          value={values.scheduler.duration}
          onChange={handleChange}
        />
      </Box>
    </>
  )
}

export default ScheduleStep
