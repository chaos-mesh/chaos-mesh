import * as Yup from 'yup'

import { FormikProps, FormikValues, getIn } from 'formik'
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
        getIn(errors, 'spec.historyLimit') && getIn(touched, 'spec.historyLimit')
          ? getIn(errors, 'spec.historyLimit')
          : T('newS.basic.historyLimitHelper')
      }
      error={getIn(errors, 'spec.historyLimit') && getIn(touched, 'spec.historyLimit') ? true : false}
    />
    <SelectField
      name="spec.concurrencyPolicy"
      label={T('newS.basic.concurrencyPolicy')}
      helperText={
        getIn(errors, 'spec.concurrencyPolicy') && getIn(touched, 'spec.concurrencyPolicy')
          ? getIn(errors, 'spec.concurrencyPolicy')
          : T('newS.basic.concurrencyPolicyHelper')
      }
      error={getIn(errors, 'spec.concurrencyPolicy') && getIn(touched, 'spec.concurrencyPolicy') ? true : false}
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
        getIn(errors, 'spec.startingDeadlineSeconds') && getIn(touched, 'spec.startingDeadlineSeconds')
          ? getIn(errors, 'spec.startingDeadlineSeconds')
          : T('newS.basic.startingDeadlineSecondsHelper')
      }
      error={
        getIn(errors, 'spec.startingDeadlineSeconds') && getIn(touched, 'spec.startingDeadlineSeconds') ? true : false
      }
    />
  </>
)

export const schema = {
  historyLimit: Yup.number().min(1, 'The historyLimit is at least 1'),
  concurrencyPolicy: Yup.string().required('The concurrencyPolicy is required'),
  startingDeadlineSeconds: Yup.number().min(0, 'The startingDeadlineSeconds is at least 0').nullable(true),
}
