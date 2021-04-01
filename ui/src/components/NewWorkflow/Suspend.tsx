import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { validateDuration } from 'lib/formikhelpers'

const Suspend = () => {
  return (
    <Paper>
      <PaperTop title={T('newW.suspendTitle')} />
      <Formik initialValues={{ duration: '' }} onSubmit={() => {}}>
        {({ errors, touched }) => (
          <Form>
            <TextField
              fast
              name="duration"
              label={T('newE.schedule.duration')}
              validate={validateDuration}
              helperText={errors.duration && touched.duration ? errors.duration : T('newW.durationHelper')}
              error={errors.duration && touched.duration ? true : false}
            />
            <Submit />
          </Form>
        )}
      </Formik>
    </Paper>
  )
}

export default Suspend
