import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import { setTemplate } from 'slices/workflows'
import { useStoreDispatch } from 'store'

export interface SuspendValues {
  name: string
  deadline: string
}

interface SuspendProps {
  initialValues?: SuspendValues
  onSubmit?: (values: SuspendValues) => void
}

const Suspend: React.FC<SuspendProps> = ({ initialValues, onSubmit }) => {
  const dispatch = useStoreDispatch()

  const defaultOnSubmit = ({ name, deadline }: SuspendValues) => {
    dispatch(
      setTemplate({
        type: 'suspend',
        name,
        deadline,
        children: [],
      })
    )
  }

  return (
    <Paper>
      <Space>
        <PaperTop title={T('newW.suspendTitle')} />
        <Formik initialValues={initialValues || { name: '', deadline: '' }} onSubmit={onSubmit || defaultOnSubmit}>
          {({ errors, touched }) => (
            <Form>
              <Space>
                <TextField
                  fast
                  name="name"
                  label={T('common.name')}
                  validate={validateName(T('newW.nameValidation') as unknown as string)}
                  helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                  error={errors.name && touched.name ? true : false}
                />
                <TextField
                  fast
                  name="deadline"
                  label={T('newW.node.deadline')}
                  validate={validateDeadline(T('newW.node.deadlineValidation') as unknown as string)}
                  helperText={errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')}
                  error={errors.deadline && touched.deadline ? true : false}
                />
              </Space>
              <Submit />
            </Form>
          )}
        </Formik>
      </Space>
    </Paper>
  )
}

export default Suspend
