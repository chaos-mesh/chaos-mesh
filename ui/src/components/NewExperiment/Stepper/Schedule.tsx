import { Box, Divider, FormControlLabel, Switch, Typography } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, makeStyles } from '@material-ui/core/styles'

import { StepperFormProps } from '../types'
import { TextField } from 'components/FormField'

const useStyles = makeStyles((theme: Theme) => ({
  subtitle: {
    margin: `${theme.spacing(4)} 0`,
  },
  description: {
    margin: `${theme.spacing(4)} 0`,
    padding: theme.spacing(4),
    background: theme.palette.background.default,
  },
}))

interface ScheduleStepProps {
  formProps: StepperFormProps
}

const ScheduleStep: React.FC<ScheduleStepProps> = ({ formProps }) => {
  const classes = useStyles()

  const [isImmediate, setIsImmediate] = useState(true)
  const { values, handleBlur, handleChange } = formProps

  const handleChecked = (event: React.ChangeEvent<HTMLInputElement>, checked: boolean) => {
    setIsImmediate(checked)
  }

  return (
    <Box maxWidth="30rem" mx="auto">
      <FormControlLabel
        control={<Switch checked={isImmediate} onChange={handleChecked} name="immediate" />}
        label="Immediate Job"
      />

      <Box hidden={isImmediate} mt={2}>
        <Divider />
        <Typography variant="subtitle2" component="h3" className={classes.subtitle}>
          Crontab Job
        </Typography>
        <Typography className={classes.description}>
          <code>
            Crontab job xxxxxxx Field name | Mandatory? | Allowed values | Allowed special characters ---------- |
            ---------- | -------------- | -------------------------- Minutes | Yes | 0-59 | * / , - Hours | Yes | 0-23 |
            * / , - Day of month | Yes | 1-31 | * / , - ? Month | Yes | 1-12 or JAN-DEC | * / , - Day of week | Yes |
            0-6 or SUN-SAT | * / , - ? Entry | Description | Equivalent To ----- | ----------- | ------------- @yearly
            (or @annually) | Run once a year, midnight, Jan. 1st | 0 0 1 1 * @monthly | Run once a month, midnight,
            first of month | 0 0 1 * * @weekly | Run once a week, midnight between Sat/Sun | 0 0 * * 0 @daily (or
            @midnight) | Run once a day, midnight | 0 0 * * * @hourly | Run once an hour, beginning of hour | 0 * * * *
          </code>
        </Typography>

        <TextField
          id="schedule.cron"
          label="Cron"
          type="text"
          autoComplete="off"
          helperText="Schedule crontab: 30 * * * *"
          value={values.schedule.cron}
          onBlur={handleBlur}
          onChange={handleChange}
        />

        <TextField
          id="schedule.duration"
          label="Duration"
          type="text"
          autoComplete="off"
          helperText="Schedule Duration: 1h30m"
          value={values.schedule.duration}
          onBlur={handleBlur}
          onChange={handleChange}
        />
      </Box>
    </Box>
  )
}

export default ScheduleStep
