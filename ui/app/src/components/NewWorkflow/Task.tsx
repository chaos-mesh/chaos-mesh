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
import { Autocomplete, Box, IconButton, TextField as MUITextField, Typography } from '@mui/material'
import { Branch, Template, TemplateType } from 'slices/workflows'
import { Form, Formik } from 'formik'
import { LabelField, Submit, TextField } from 'components/FormField'
import { useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'
import { validateImage, validateName } from 'lib/formikhelpers'

import Add from './Add'
import AddCircleIcon from '@mui/icons-material/AddCircle'
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown'
import ArrowRightIcon from '@mui/icons-material/ArrowRight'
import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import RemoveCircleIcon from '@mui/icons-material/RemoveCircle'
import Space from '@ui/mui-extends/esm/Space'
import i18n from 'components/T'
import { resetNewExperiment } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'

interface TaskProps extends FormProps {
  childrenCount: number
  submitTemplate: (template: Template) => void
  templates: Template[]
}

interface FormProps {
  name?: string
  deadline?: string
  type: TemplateType
  container: Container
  conditionalBranches: Branch[]
}
interface Container {
  name: string
  image: string
  command: string[]
}

const Task: React.FC<TaskProps> = (props) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()
  const formRef = useRef<any>()

  const { templates: storeTemplates } = useStoreSelector((state) => state.workflows)
  const [templates, setTemplates] = useState<Template[]>(props.templates)
  const templateNames = [...new Set([...storeTemplates, ...templates].map((t) => t.name))]

  const submitTask = () => {
    const { name, deadline, container, conditionalBranches } = formRef.current.values
    const template: Template = {
      type: props.type,
      name: name.trim(),
      deadline,
      children: templates,
      custom: {
        container,
        conditionalBranches,
      },
    }
    props.submitTemplate(template)
  }

  // expand is an int index, stands for the detail page of (expand)-th child task is expanded
  // so it's obvious that there is only one expanded detail page at a time
  // when expand is -1, means no detail page is expanded
  const [expand, setExpand] = useState(-1)
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
    <>
      <Formik
        innerRef={formRef}
        initialValues={
          {
            name: props.name || '',
            deadline: props.deadline || '',
            type: props.type,
            container: props.container,
            conditionalBranches: props.conditionalBranches,
          } as FormProps
        }
        onSubmit={submitTask}
      >
        {({ values, setFieldValue, errors, touched }) => {
          const { conditionalBranches } = values

          const addBranch = (branch: Branch) => () => {
            if (branch.target === '') {
              dispatch(
                setAlert({
                  type: 'warning',
                  message: i18n('newW.messages.m2', intl),
                })
              )

              return
            }

            setFieldValue(
              'conditionalBranches',
              conditionalBranches.concat([
                {
                  target: '',
                  expression: '',
                },
              ])
            )
            // setChildrenCount(childrenCount + 1)
          }

          const removeBranch = (index: number) => () => {
            setFieldValue(
              'conditionalBranches',
              conditionalBranches.filter((_: any, i: number) => index !== i)
            )
            // setChildrenCount(childrenCount - 1)
            setTemplates(templates.filter((_: any, i: number) => index !== i))
          }

          const conditionalBranchTargetSelected =
            (index: number) => (_: any, newVal: string | null, reason: string) => {
              const name = `conditionalBranches[${index}].target`

              if (reason === 'clear') {
                setFieldValue(name, '')

                return
              }

              setFieldValue(name, newVal)

              if (templateNames.includes(newVal!)) {
                const template = [...storeTemplates, ...templates].find((t) => t.name === newVal)!

                const tmp = JSON.parse(JSON.stringify(templates))
                tmp[index] = template

                setTemplates(tmp)
                // setNum(tmp.length)
              }
            }

          return (
            props.type === 'custom' && (
              <Box mt={3} ml={8}>
                <Form>
                  <Paper>
                    <PaperTop title={i18n(`newW.${values.type}Title`)} boxProps={{ mb: 3 }} />
                    <Space>
                      <TextField
                        fast
                        name="name"
                        label={i18n('common.name')}
                        validate={validateName(i18n('newW.node.nameValidation', intl))}
                        helperText={errors.name && touched.name ? errors.name : i18n('newW.node.nameHelper')}
                        error={errors.name && touched.name ? true : false}
                      />
                      <Typography variant="body2">{i18n('newW.node.container.title')}</Typography>
                      <TextField
                        fast
                        name="container.name"
                        label={i18n('common.name')}
                        validate={validateName(i18n('newW.node.container.nameValidation', intl))}
                        helperText={
                          errors.container?.name && touched.container?.name
                            ? errors.container.name
                            : i18n('newW.node.container.nameHelper')
                        }
                        error={errors.container?.name && touched.container?.name ? true : false}
                      />
                      <TextField
                        fast
                        name="container.image"
                        label={i18n('newW.node.container.image')}
                        validate={validateImage(i18n('newW.node.container.imageValidation', intl))}
                        helperText={
                          errors.container?.image && touched.container?.image
                            ? errors.container.image
                            : i18n('newW.node.container.imageHelper')
                        }
                        error={errors.container?.image && touched.container?.image ? true : false}
                      />
                      <LabelField
                        name="container.command"
                        label={i18n('newW.node.container.command')}
                        helperText={i18n('newW.node.container.commandHelper')}
                      />
                      <Typography variant="body2">{i18n('newW.node.conditionalBranches.title')}</Typography>
                      {conditionalBranches.length > 0 &&
                        conditionalBranches.map((d, i) => (
                          <Space key={i} direction="row" alignItems="center">
                            <Typography component="div" variant="button">
                              if
                            </Typography>
                            <TextField
                              name={`conditionalBranches[${i}].expression`}
                              label={i18n('newW.node.conditionalBranches.expression')}
                            />
                            <Typography component="div" variant="button">
                              then
                            </Typography>
                            <Autocomplete
                              sx={{ width: 360 }}
                              options={templateNames}
                              noOptionsText={i18n('common.noOptions')}
                              value={(function () {
                                if (templates[i] && templates[i].name !== conditionalBranches[i].target) {
                                  const name = templates[i].name

                                  setFieldValue(`conditionalBranches[${i}].target`, name)

                                  return name
                                }

                                return conditionalBranches[i].target
                              })()}
                              onChange={conditionalBranchTargetSelected(i)}
                              renderInput={(params) => (
                                <MUITextField
                                  {...params}
                                  name={`conditionalBranches[${i}].target`}
                                  label={i18n('newW.node.conditionalBranches.target')}
                                  size="small"
                                  fullWidth
                                />
                              )}
                              PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
                            />
                            {i !== conditionalBranches.length - 1 && (
                              <IconButton color="secondary" size="small" onClick={removeBranch(i)}>
                                <RemoveCircleIcon />
                              </IconButton>
                            )}
                            {i === conditionalBranches.length - 1 && (
                              <IconButton color="primary" size="small" onClick={addBranch(d)}>
                                <AddCircleIcon />
                              </IconButton>
                            )}
                          </Space>
                        ))}
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
          )
        }}
      </Formik>
    </>
  )
}

export default Task
