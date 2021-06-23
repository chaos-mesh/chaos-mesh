import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { Template, setTemplate, updateTemplate } from 'slices/workflows'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import { useStoreDispatch } from 'store'

export interface SuspendValues {
  name: string
  deadline: string
}

interface SuspendProps {
  initialValues?: SuspendValues
  update?: number
  updateCallback?: () => void
}

const Suspend: React.FC<SuspendProps> = ({ initialValues, update, updateCallback }) => {
  const dispatch = useStoreDispatch()

  const onSubmit = ({ name, deadline }: SuspendValues) => {
    const template: Template = {
      type: 'suspend',
      name,
      deadline,
    }

    dispatch(update !== undefined ? updateTemplate({ ...template, index: update }) : setTemplate(template))
    typeof updateCallback === 'function' && updateCallback()
  }

  return (
    <Paper>
      <Space>
        <PaperTop title={T('newW.suspendTitle')} />
        <Formik initialValues={initialValues || { name: '', deadline: '' }} onSubmit={onSubmit}>
          {({ errors, touched }) => (
            <Form>
              <Space>
                <TextField
                  fast
                  name="name"
                  label={T('common.name')}
                  validate={validateName(T('newW.node.nameValidation') as unknown as string)}
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
