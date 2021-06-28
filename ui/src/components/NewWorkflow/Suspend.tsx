import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import { Template } from 'slices/workflows'
import { schemaBasic } from './types'

export interface SuspendValues {
  name: string
  deadline: string
}

interface SuspendProps {
  initialValues?: SuspendValues
  submit: (template: Template) => void
}

const Suspend: React.FC<SuspendProps> = ({ initialValues, submit }) => {
  const onSubmit = ({ name, deadline }: SuspendValues) => {
    const values = schemaBasic.cast({ name, deadline })

    submit({
      type: 'suspend',
      ...values!,
    })
  }

  return (
    <Paper>
      <Space>
        <PaperTop title={T('newW.suspendTitle')} />
        <Formik
          initialValues={initialValues || { name: '', deadline: '' }}
          validationSchema={schemaBasic}
          onSubmit={onSubmit}
        >
          {({ errors, touched }) => (
            <Form>
              <Space>
                <TextField
                  fast
                  name="name"
                  label={T('common.name')}
                  helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                  error={errors.name && touched.name ? true : false}
                />
                <TextField
                  fast
                  name="deadline"
                  label={T('newW.node.deadline')}
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
