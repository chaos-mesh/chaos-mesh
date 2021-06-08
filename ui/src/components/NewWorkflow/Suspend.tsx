import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { validateDuration, validateName } from 'lib/formikhelpers'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import { setTemplate } from 'slices/workflows'
import { useStoreDispatch } from 'store'

export interface SuspendValues {
  name: string
  duration: string
}

interface SuspendProps {
  initialValues?: SuspendValues
  onSubmit?: (values: SuspendValues) => void
}

const Suspend: React.FC<SuspendProps> = ({ initialValues, onSubmit }) => {
  const dispatch = useStoreDispatch()

  const defaultOnSubmit = ({ name, duration }: SuspendValues) => {
    dispatch(
      setTemplate({
        type: 'suspend',
        name,
        duration,
        experiments: [],
      })
    )
  }

  return (
    <Paper>
      <PaperTop title={T('newW.suspendTitle')} />
      <Formik initialValues={initialValues || { name: '', duration: '' }} onSubmit={onSubmit || defaultOnSubmit}>
        {({ errors, touched }) => (
          <Form>
            <Space>
              <TextField
                name="name"
                label={T('common.name')}
                validate={validateName(T('newW.nameValidation') as unknown as string)}
                helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                error={errors.name && touched.name ? true : false}
              />
              <TextField
                fast
                name="duration"
                label={T('newE.run.duration')}
                validate={validateDuration(T('newW.durationValidation') as unknown as string)}
                helperText={errors.duration && touched.duration ? errors.duration : T('newW.node.durationHelper')}
                error={errors.duration && touched.duration ? true : false}
              />
              <Submit />
            </Space>
          </Form>
        )}
      </Formik>
    </Paper>
  )
}

export default Suspend
