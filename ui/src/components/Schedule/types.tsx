import * as Yup from 'yup'

import { FormikProps, FormikValues } from 'formik'
import { InputAdornment, Link, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

import { FormattedMessage } from 'react-intl'
import T from 'components/T'

export interface ScheduleSpecific {
  schedule: string
  starting_deadline_seconds?: number
  concurrency_policy?: 'Forbid' | 'Allow'
  history_limit?: number
}

export const data: ScheduleSpecific = {
  schedule: '',
  starting_deadline_seconds: undefined,
  concurrency_policy: 'Forbid',
  history_limit: 1,
}

export const Fields = ({ errors, touched }: Pick<FormikProps<FormikValues>, 'errors' | 'touched'>) => (
  <>
    <TextField
      fast
      name="schedule"
      label={T('schedules.single')}
      helperText={
        errors.schedule && touched.schedule ? (
          errors.schedule
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
      error={errors.schedule && touched.schedule ? true : false}
    />
    <TextField
      fast
      type="number"
      name="starting_deadline_seconds"
      label={T('newS.basic.startingDeadlineSeconds')}
      InputProps={{
        endAdornment: <InputAdornment position="end">{T('common.seconds')}</InputAdornment>,
      }}
      helperText={
        errors.starting_deadline_seconds && touched.starting_deadline_seconds
          ? errors.starting_deadline_seconds
          : T('newS.basic.startingDeadlineSecondsHelper')
      }
      error={errors.starting_deadline_seconds && touched.starting_deadline_seconds ? true : false}
    />
    <SelectField
      name="concurrency_policy"
      label={T('newS.basic.concurrencyPolicy')}
      helperText={
        errors.concurrency_policy && touched.concurrency_policy
          ? errors.concurrency_policy
          : T('newS.basic.concurrencyPolicyHelper')
      }
      error={errors.concurrency_policy && touched.concurrency_policy ? true : false}
    >
      <MenuItem value="Forbid">{T('newS.basic.forbid')}</MenuItem>
      <MenuItem value="Allow">{T('newS.basic.allow')}</MenuItem>
    </SelectField>
    <TextField
      fast
      type="number"
      name="history_limit"
      label={T('newS.basic.historyLimit')}
      helperText={
        errors.history_limit && touched.history_limit ? errors.history_limit : T('newS.basic.historyLimitHelper')
      }
      error={errors.history_limit && touched.history_limit ? true : false}
    />
  </>
)

export const schema = {
  schedule: Yup.string().required('The schedule is required'),
  starting_deadline_seconds: Yup.number().min(0, 'The startingDeadlineSeconds is at least 0'),
  history_limit: Yup.number().min(1, 'The historyLimit is at least 1'),
}
