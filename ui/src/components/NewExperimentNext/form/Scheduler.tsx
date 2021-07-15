import { Box, FormControlLabel, Switch, Typography } from '@material-ui/core'
import { FormikErrors, FormikTouched, getIn, useFormikContext } from 'formik'
import { useEffect, useState } from 'react'

import T from 'components/T'
import { TextField } from 'components/FormField'
import { useStoreSelector } from 'store'
import { validateDuration } from 'lib/formikhelpers'

interface SchedulerProps {
  errors: FormikErrors<Record<string, any>>
  touched: FormikTouched<Record<string, any>>
  inSchedule?: boolean
}

const Scheduler: React.FC<SchedulerProps> = ({ errors, touched, inSchedule = false }) => {
  const { fromExternal, basic } = useStoreSelector((state) => state.experiments)
  const { values, setFieldValue } = useFormikContext()

  const [continuous, setContinuous] = useState(false)

  useEffect(() => {
    if (!inSchedule && fromExternal && basic.spec.duration === '') {
      setContinuous(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fromExternal])

  const handleChecked = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked

    setContinuous(checked)

    if (checked && getIn(values, 'spec.duration') !== '') {
      setFieldValue('spec.duration', '')
    }
  }

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography>{T('newE.steps.run')}</Typography>
        {!inSchedule && (
          <FormControlLabel
            style={{ marginRight: 0 }}
            control={
              <Switch name="continuous" color="primary" size="small" checked={continuous} onChange={handleChecked} />
            }
            label={T('newE.run.continuous')}
          />
        )}
      </Box>

      {!continuous && (
        <TextField
          fast
          name="spec.duration"
          label={T('common.duration')}
          validate={validateDuration()}
          helperText={
            getIn(errors, 'spec.duration') && getIn(touched, 'spec.duration')
              ? getIn(errors, 'spec.duration')
              : T('common.durationHelper')
          }
          error={getIn(errors, 'spec.duration') && getIn(touched, 'spec.duration') ? true : false}
        />
      )}
    </>
  )
}

export default Scheduler
