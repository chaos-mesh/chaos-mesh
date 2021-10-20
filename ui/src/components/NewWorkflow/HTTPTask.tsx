import { Form, Formik } from 'formik'
import { FormControlLabel, Switch } from '@material-ui/core'
import { Submit, TextField } from 'components/FormField'
import { Template, TemplateType } from 'slices/workflows'
import { parseHTTPTask, renderHTTPTask } from 'api/workflows'
import { useEffect, useRef, useState } from 'react'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import { RequestForm } from 'api/workflows.type'
import Space from 'components-mui/Space'
import T from 'components/T'

interface HTTPTaskProps extends CommonTemplateProps {
  childrenCount: number
  submitTemplate: (template: Template) => void
}

interface CommonTemplateProps {
  name?: string
  deadline?: string
  type: TemplateType
  templates: Template[]
  externalTemplate?: Template
}

interface FromProps {
  name: string
  url: string
  method: string
  body: string
  follow: boolean
  json: boolean
}

const HTTPTask: React.FC<HTTPTaskProps> = (props) => {
  const { submitTemplate } = props
  const onSubmit = (form: RequestForm) => {
    renderHTTPTask(form)
      .then((response) => {
        const { name, task } = response.data!
        const { container } = task
        const result: Template = {
          name,
          children: [],
          type: TemplateType.Custom,
          custom: {
            container,
            conditionalBranches: [],
          },
        }
        submitTemplate(result)
      })
      .catch(console.error)
  }
  const formRef = useRef<any>()
  const [initialValues, setInitialValues] = useState<FromProps>({
    name: props.name || '',
    url: '',
    method: '',
    body: '',
    follow: false,
    json: false,
  })

  const validateRequestForm = (newValue: RequestForm) => {
    console.log(newValue)
    const errors: any = {}
    return errors
  }

  useEffect(() => {
    if (props.externalTemplate) {
      parseHTTPTask({
        name: props.externalTemplate.name,
        type: props.externalTemplate.type,
        task: {
          container: props.externalTemplate.custom!.container,
          conditionalBranches: props.externalTemplate.custom!.conditionalBranches,
        },
      })
        .then((response) => {
          if (response.data) {
            const parsedForm = response.data as RequestForm
            setInitialValues({
              name: parsedForm.name,
              url: parsedForm.url,
              method: parsedForm.method,
              body: parsedForm.body,
              follow: parsedForm.follow || false,
              json: parsedForm.json || false,
            })
          }
        })
        .catch(console.error)
    }
    return () => {}
  }, [props.externalTemplate])

  return (
    <>
      <Paper>
        <Space>
          <PaperTop title={T('newW.httpTitle')} />

          <Formik
            innerRef={formRef}
            initialValues={initialValues}
            enableReinitialize
            onSubmit={onSubmit}
            validate={validateRequestForm}
            validateOnBlur={false}
          >
            {({ values, errors, touched, handleChange, handleBlur, handleSubmit, isSubmitting }) => {
              return (
                <Form>
                  <Space>
                    <TextField
                      name="name"
                      label={T('common.name')}
                      //   validate={validateName(T('newW.node.nameValidation', intl))}
                      helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                      error={errors.name && touched.name ? true : false}
                      size="small"
                      fullWidth
                    />
                    <TextField
                      name="url"
                      label={T('newW.node.httpRequest.url')}
                      //   validate={validateDeadline(T('newW.node.deadlineValidation', intl))}
                      //   helperText={errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')}
                      //   error={errors.deadline && touched.deadline ? true : false}
                      size="small"
                      fullWidth
                    />
                    <TextField name="method" label={T('newW.node.httpRequest.method')} size="small" fullWidth />
                    <TextField name="body" label={T('newW.node.httpRequest.body')} size="small" fullWidth />

                    <FormControlLabel
                      style={{ marginRight: 0 }}
                      label={T('newW.node.httpRequest.follow')}
                      control={<Switch name="follow" onChange={handleChange} />}
                    />

                    <FormControlLabel
                      style={{ marginRight: 0 }}
                      label={T('newW.node.httpRequest.json')}
                      control={<Switch name="json" onChange={handleChange} />}
                    />
                    <Submit />
                  </Space>
                </Form>
              )
            }}
          </Formik>
        </Space>
      </Paper>
    </>
  )
}

export default HTTPTask
