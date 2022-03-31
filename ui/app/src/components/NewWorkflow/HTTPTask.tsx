/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Form, Formik } from 'formik'
import { FormControlLabel, Switch } from '@mui/material'
import { MenuItem, Typography } from '@mui/material'
import { SelectField, Submit, TextField } from 'components/FormField'
import { Template, TemplateType } from 'slices/workflows'
import { parseHTTPTask, renderHTTPTask } from 'api/workflows'
import { useEffect, useRef, useState } from 'react'

import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import { RequestForm } from 'api/workflows.type'
import Space from '@ui/mui-extends/esm/Space'
import i18n from 'components/T'
import { makeStyles } from '@mui/styles'
import { useIntl } from 'react-intl'
import { validateName } from 'lib/formikhelpers'

const useStyles = makeStyles({
  field: {
    width: 180,
  },
})

const HTTPMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTION']

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

const HTTPTask: React.FC<HTTPTaskProps> = (props) => {
  const intl = useIntl()
  const classes = useStyles()

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
  const [initialValues, setInitialValues] = useState<RequestForm>({
    name: props.name || '',
    url: '',
    method: '',
    body: '',
    followLocation: false,
    jsonContent: false,
  })

  useEffect(() => {
    if (props.externalTemplate) {
      // TODO: use unified name
      const backendType = props.externalTemplate.type === 'custom' ? 'Task' : props.externalTemplate.type
      parseHTTPTask({
        name: props.externalTemplate.name,
        templateType: backendType,
        task: {
          container: props.externalTemplate.custom!.container,
          conditionalBranches: props.externalTemplate.custom!.conditionalBranches,
        },
      })
        .then((response) => {
          if (response.data) {
            const parsedForm = response.data as RequestForm
            setInitialValues({
              ...parsedForm,
              followLocation: parsedForm.followLocation || false,
              jsonContent: parsedForm.jsonContent || false,
            })
          }
        })
        .catch(console.error)
    }
    return () => {}
  }, [props.externalTemplate])

  return (
    <Paper>
      <Space>
        <PaperTop title={i18n('newW.httpTitle')} />
        <Formik
          innerRef={formRef}
          initialValues={initialValues}
          enableReinitialize
          onSubmit={onSubmit}
          validateOnBlur={false}
        >
          {({ values, errors, touched, handleChange }) => {
            return (
              <Form>
                <Space>
                  <TextField
                    name="name"
                    label={i18n('common.name')}
                    validate={validateName(i18n('newW.node.nameValidation', intl))}
                    helperText={errors.name && touched.name ? errors.name : i18n('newW.node.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                    size="small"
                    fullWidth
                  />
                  <TextField
                    name="url"
                    label={i18n('newW.node.httpRequest.url')}
                    helperText={errors.url && touched.url ? errors.url : i18n('newW.node.httpRequest.urlHelper')}
                    error={errors.url && touched.url ? true : false}
                    size="small"
                    fullWidth
                  />

                  <SelectField
                    className={classes.field}
                    name="method"
                    label={i18n('newW.node.httpRequest.method')}
                    helperText={
                      errors.method && touched.method ? errors.method : i18n('newW.node.httpRequest.methodHelper')
                    }
                    size="small"
                  >
                    {HTTPMethods.map((method) => (
                      <MenuItem key={method} value={method}>
                        <Typography variant="body2">{method}</Typography>
                      </MenuItem>
                    ))}
                  </SelectField>
                  {(values.method === 'POST' || values.method === 'PUT') && (
                    <TextField
                      name="body"
                      label={i18n('newW.node.httpRequest.body')}
                      helperText={errors.body && touched.body ? errors.body : i18n('newW.node.httpRequest.bodyHelper')}
                      size="small"
                      fullWidth
                    />
                  )}

                  <FormControlLabel
                    style={{ marginRight: 0 }}
                    label={i18n('newW.node.httpRequest.follow')}
                    control={<Switch name="followLocation" onChange={handleChange} />}
                  />

                  <FormControlLabel
                    style={{ marginRight: 0 }}
                    label={i18n('newW.node.httpRequest.json')}
                    control={<Switch name="jsonContent" onChange={handleChange} />}
                  />
                  <Submit />
                </Space>
              </Form>
            )
          }}
        </Formik>
      </Space>
    </Paper>
  )
}

export default HTTPTask
