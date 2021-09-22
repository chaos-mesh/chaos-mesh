import { FormControlLabel, Switch, TextField } from '@material-ui/core'
import { Template, TemplateType } from 'slices/workflows'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import { Submit } from 'components/FormField'
import T from 'components/T'
import { useFormik } from 'formik'

export interface RenderedHTTPTask {
  name: string
  container: {
    name: string
    image: string
    command: string[]
  }
}

export interface RequestFlags {
  name: string
  url: string
  method: string
  body: string
  follow: boolean
  json: boolean
}

interface HTTPTaskProps {
  initialValues?: RenderedHTTPTask
  submit: (template: Template) => void
}

const HTTPTask: React.FC<HTTPTaskProps> = ({ initialValues, submit }) => {
  const onSubmit = ({ name }: RequestFlags) => {
    const values = { name }
    // TODO: convert RequestFlags to RenderedHTTPTask
    submit({
      type: TemplateType.HTTP,
      ...values!,
    })
  }

  const validateRequestForm = (newValue: RequestFlags) => {
    console.log(newValue)
  }
  const formik = useFormik({
    initialValues: { name: '', url: '', method: 'GET', body: '', follow: false, json: false },
    validate: validateRequestForm,
    validateOnBlur: false,
    onSubmit: onSubmit,
  })

  return (
    <>
      <Paper>
        <Space>
          <PaperTop title={T('newW.httpTitle')} />
          <form onSubmit={formik.handleSubmit}>
            <Space>
              <TextField
                name="name"
                label={T('common.name')}
                //   validate={validateName(T('newW.node.nameValidation', intl))}
                helperText={formik.errors.name && formik.touched.name ? formik.errors.name : T('newW.node.nameHelper')}
                error={formik.errors.name && formik.touched.name ? true : false}
                onChange={formik.handleChange}
                size="small"
                fullWidth
              />
              <TextField
                name="url"
                label={T('newW.node.httpRequest.url')}
                //   validate={validateDeadline(T('newW.node.deadlineValidation', intl))}
                //   helperText={errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')}
                //   error={errors.deadline && touched.deadline ? true : false}
                onChange={formik.handleChange}
                size="small"
                fullWidth
              />
              <TextField
                name="method"
                label={T('newW.node.httpRequest.method')}
                onChange={formik.handleChange}
                size="small"
                fullWidth
              />
              <TextField
                name="body"
                label={T('newW.node.httpRequest.body')}
                onChange={formik.handleChange}
                size="small"
                fullWidth
              />

              <FormControlLabel
                style={{ marginRight: 0 }}
                label={T('newW.node.httpRequest.follow')}
                control={<Switch name="follow" onChange={formik.handleChange} />}
              />

              <FormControlLabel
                style={{ marginRight: 0 }}
                label={T('newW.node.httpRequest.json')}
                control={<Switch name="json" onChange={formik.handleChange} />}
              />
            </Space>
            <Submit />
          </form>
        </Space>
      </Paper>
    </>
  )
}

export default HTTPTask
