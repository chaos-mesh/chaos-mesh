import { Box, IconButton, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { Template, TemplateType } from 'slices/workflows'
import { useRef, useState } from 'react'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import Add from './Add'
import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowRightIcon from '@material-ui/icons/ArrowRight'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import React from 'react'
import Space from 'components-mui/Space'
import T from 'components/T'
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
          message: T('newW.messages.m1', intl),
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
    <>
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
                  <PaperTop title={T(`newW.${values.type}Title`)} boxProps={{ mb: 3 }} />
                  <Space direction="row">
                    <TextField
                      name="name"
                      label={T('common.name')}
                      validate={validateName(T('newW.nameValidation', intl))}
                      helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                      error={errors.name && touched.name ? true : false}
                    />
                    <TextField
                      name="deadline"
                      label={T('newW.node.deadline')}
                      validate={validateDeadline(T('newW.node.deadlineValidation', intl))}
                      helperText={errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')}
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
                            : `${T('newW.node.child', intl)} ${index + 1}`}
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
    </>
  )
}
export default SerialOrParallel
