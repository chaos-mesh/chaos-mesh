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
import { Box, IconButton, Typography } from '@mui/material'
import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { Template, TemplateType } from 'slices/workflows'
import { useRef, useState } from 'react'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import Add from './Add'
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown'
import ArrowRightIcon from '@mui/icons-material/ArrowRight'
import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import React from 'react'
import Space from '@ui/mui-extends/esm/Space'
import i18n from 'components/T'
import { resetNewExperiment } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

interface SerialOrParallelProps extends FormProps {
  childrenCount: number
  submitTemplate: (template: Template) => void
  templates: Template[]
}
interface FormProps {
  name?: string
  deadline?: string
  type: TemplateType
}
/**
 * SerialOrParallel component is the editor of workflow template with type Serial or Parallel
 * @param props SerialOrParallelProps
 * @returns
 */
const SerialOrParallel: React.FC<SerialOrParallelProps> = (props) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const formRef = useRef<any>()

  // expand is an int index, stands for the detail page of (expand)-th child task is expanded
  // so it's obvious that there is only one expanded detail page at a time
  // when expand is -1, means no detail page is expanded
  const [expand, setExpand] = useState(-1)

  const [templates, setTemplates] = useState<Template[]>(props.templates || [])

  const submitSerialOrParallel = () => {
    const { name, deadline } = formRef.current.values
    const template: Template = {
      type: props.type,
      name: name.trim(),
      deadline,
      children: templates,
    }
    props.submitTemplate(template)
  }

  const onValidate = (values: FormProps) => {
    const errors: any = {}
    return errors
  }

  const switchExpand = (index: number) => () => {
    if (index > templates.length) {
      dispatch(
        setAlert({
          type: 'warning',
          // Please fill in the current branch first
          message: i18n('newW.messages.m1', intl),
        })
      )

      return
    }

    setExpand(
      expand === index
        ? (function () {
            dispatch(resetNewExperiment())

            return -1
          })()
        : index
    )
  }

  return (
    <Formik
      innerRef={formRef}
      initialValues={
        {
          name: props.name || '',
          deadline: props.deadline || '',
          type: props.type,
        } as FormProps
      }
      enableReinitialize
      onSubmit={submitSerialOrParallel}
      validate={onValidate}
      validateOnBlur={false}
    >
      {({ values, setFieldValue, errors, touched }) => {
        return (
          <Box mt={3} ml={8}>
            <Form>
              <Paper>
                <PaperTop title={i18n(`newW.${values.type}Title`)} boxProps={{ mb: 3 }} />
                <Space direction="row">
                  <TextField
                    name="name"
                    label={i18n('common.name')}
                    validate={validateName(i18n('newW.nameValidation', intl))}
                    helperText={errors.name && touched.name ? errors.name : i18n('newW.node.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />
                  <TextField
                    name="deadline"
                    label={i18n('newW.node.deadline')}
                    validate={validateDeadline(i18n('newW.node.deadlineValidation', intl))}
                    helperText={
                      errors.deadline && touched.deadline ? errors.deadline : i18n('newW.node.deadlineHelper')
                    }
                    error={errors.deadline && touched.deadline ? true : false}
                  />
                </Space>
                <Submit disabled={templates.length !== props.childrenCount} />
              </Paper>
            </Form>

            {Array(props.childrenCount)
              .fill(0)
              .map((_, index) => (
                <Box key={index} ml={8}>
                  <Paper
                    sx={{
                      my: 6,
                      p: 1.5,
                      borderColor: templates[index] ? 'success.main' : undefined,
                    }}
                  >
                    <Box display="flex" alignItems="center">
                      <IconButton size="small" onClick={switchExpand(index)}>
                        {expand === index ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                      </IconButton>
                      <Typography component="div" sx={{ ml: 1 }}>
                        {templates.length > index
                          ? templates[index].name
                          : `${i18n('newW.node.child', intl)} ${index + 1}`}
                      </Typography>
                    </Box>
                  </Paper>
                  {expand === index && (
                    <Box mt={6}>
                      <Add
                        childIndex={index}
                        parentTemplates={templates}
                        setParentTemplates={setTemplates}
                        setParentExpand={setExpand}
                        externalTemplate={templates[index]}
                      />
                    </Box>
                  )}
                </Box>
              ))}
          </Box>
        )
      }}
    </Formik>
  )
}
export default SerialOrParallel
