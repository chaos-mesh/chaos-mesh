import * as Yup from 'yup'

import { FormikProps, FormikValues } from 'formik'
import { InputAdornment, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

import T from 'components/T'

export interface ScheduleSpecific {
  schedule: string
  historyLimit?: number
  concurrencyPolicy?: 'Forbid' | 'Allow'
  startingDeadlineSeconds?: number
}

export const data: ScheduleSpecific = {
  schedule: '',
  historyLimit: 1,
  concurrencyPolicy: 'Forbid',
  startingDeadlineSeconds: undefined,
}

export const Fields = ({ errors, touched }: Pick<FormikProps<FormikValues>, 'errors' | 'touched'>) => (
  <>
    <TextField
      fast
      type="number"
      name="spec.historyLimit"
      label={T('newS.basic.historyLimit')}
      helperText={
        errors.historyLimit && touched.historyLimit ? errors.historyLimit : T('newS.basic.historyLimitHelper')
      }
      error={errors.historyLimit && touched.historyLimit ? true : false}
    />
    <SelectField
      name="spec.concurrencyPolicy"
      label={T('newS.basic.concurrencyPolicy')}
      helperText={
        errors.concurrencyPolicy && touched.concurrencyPolicy
          ? errors.concurrencyPolicy
          : T('newS.basic.concurrencyPolicyHelper')
      }
      error={errors.concurrencyPolicy && touched.concurrencyPolicy ? true : false}
    >
      <MenuItem value="Forbid">{T('newS.basic.forbid')}</MenuItem>
      <MenuItem value="Allow">{T('newS.basic.allow')}</MenuItem>
    </SelectField>
    <TextField
      fast
      type="number"
      name="spec.startingDeadlineSeconds"
      label={T('newS.basic.startingDeadlineSeconds')}
      InputProps={{
        endAdornment: <InputAdornment position="end">{T('common.seconds')}</InputAdornment>,
      }}
      helperText={
        errors.startingDeadlineSeconds && touched.startingDeadlineSeconds
          ? errors.startingDeadlineSeconds
          : T('newS.basic.startingDeadlineSecondsHelper')
      }
      error={errors.startingDeadlineSeconds && touched.startingDeadlineSeconds ? true : false}
    />
  </>
)

export const schema = {
  schedule: Yup.string().required('The schedule is required'),
  historyLimit: Yup.number().min(1, 'The historyLimit is at least 1'),
  startingDeadlineSeconds: Yup.number().min(0, 'The startingDeadlineSeconds is at least 0'),
}
