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
import loadable from '@loadable/component'
import CheckIcon from '@mui/icons-material/Check'
import PublishIcon from '@mui/icons-material/Publish'
import RemoveIcon from '@mui/icons-material/Remove'
import UndoIcon from '@mui/icons-material/Undo'
import {
  Box,
  Button,
  Chip,
  Grid,
  IconButton,
  ListItemIcon,
  MenuItem,
  Step,
  StepLabel,
  Stepper,
  Typography,
} from '@mui/material'
import { makeStyles } from '@mui/styles'
import { Ace } from 'ace-builds'
import { Stale } from 'api/queryUtils'
import { Form, Formik } from 'formik'
import yaml from 'js-yaml'
import _ from 'lodash'
import { useGetCommonChaosAvailableNamespaces, usePostWorkflows } from 'openapi'
import { useEffect, useState } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import Menu from '@ui/mui-extends/esm/Menu'
import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { resetNewExperiment } from 'slices/experiments'
import { setAlert, setConfirm } from 'slices/globalStatus'
import { Template, deleteTemplate, resetWorkflow } from 'slices/workflows'

import { SelectField, TextField } from 'components/FormField'
import i18n from 'components/T'

import { validateDeadline, validateName } from 'lib/formikhelpers'
import { constructWorkflow } from 'lib/formikhelpers'

import Add from './Add'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  leftSticky: {
    position: 'sticky',
    top: 0,
    height: `calc(100vh - 56px - ${theme.spacing(9)})`,
  },
  field: {
    width: 180,
    marginTop: 0,
    [theme.breakpoints.up('sm')]: {
      marginBottom: 0,
    },
    '& .MuiInputBase-input': {
      padding: 8,
    },
    '& .MuiInputLabel-root, fieldset': {
      fontSize: theme.typography.body2.fontSize,
      lineHeight: 0.875,
    },
  },
}))

type IStep = Template

export type WorkflowBasic = {
  name: string
  namespace: string
  deadline: string
}

const NewWorkflow = () => {
  const classes = useStyles()
  const intl = useIntl()
  const navigate = useNavigate()

  const state = useStoreSelector((state) => state)
  const { templates } = state.workflows
  const dispatch = useStoreDispatch()

  const [steps, setSteps] = useState<IStep[]>([])
  const [restoreIndex, setRestoreIndex] = useState(-1)
  const [workflowBasic, setWorkflowBasic] = useState<WorkflowBasic>({
    name: '',
    namespace: '',
    deadline: '',
  })
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })
  const { mutateAsync } = usePostWorkflows()

  useEffect(() => {
    return () => {
      dispatch(resetNewExperiment())
    }
  }, [dispatch])

  useEffect(() => {
    setSteps(_.isEmpty(templates) ? [] : templates)
  }, [templates])

  const resetRestore = () => {
    setRestoreIndex(-1)
  }

  const restoreExperiment = (index: number) => () => {
    if (restoreIndex !== -1) {
      resetRestore()
    } else {
      setRestoreIndex(index)
    }
  }

  const removeExperiment = (index: number) => {
    dispatch(deleteTemplate(index))
    dispatch(
      setAlert({
        type: 'success',
        message: i18n('confirm.success.delete', intl) as string,
      })
    )
    resetRestore()
  }

  const handleSelect = (name: string, index: number, action: string) => () => {
    switch (action) {
      case 'delete':
        dispatch(
          setConfirm({
            index,
            title: `${i18n('common.delete', intl)} ${name}`,
            description: i18n('newW.node.deleteDesc', intl) as string,
            handle: handleAction(action, index),
          })
        )
        break
    }
  }

  const handleAction = (action: string, index: number) => () => {
    switch (action) {
      case 'delete':
        removeExperiment(index)
        break
    }
  }

  const updateTemplateCallback = () => {
    setRestoreIndex(-1)
    dispatch(resetNewExperiment())
  }

  const onValidate = setWorkflowBasic

  const submitWorkflow = () => {
    const workflow = yamlEditor?.getValue()!

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug workflow:', workflow)
    }

    mutateAsync({
      data: yaml.load(workflow) as any,
    })
      .then(() => {
        dispatch(resetWorkflow())

        navigate('/workflows')
      })
      .catch(console.error)
  }

  return (
    <Grid container spacing={9}>
      <Grid item xs={12} md={8}>
        <Space spacing={6}>
          <Typography>{i18n('common.process')}</Typography>
          <Stepper orientation="vertical" sx={{ mt: -1, p: 0 }}>
            {steps.length > 0 &&
              steps.map((step, index) => (
                <Step key={step.name}>
                  {restoreIndex !== index ? (
                    <StepLabel icon={<CheckIcon sx={{ color: 'success.main' }} />}>
                      <Paper sx={{ p: 3, borderColor: 'success.main' }}>
                        <Box display="flex" justifyContent="space-between">
                          <Space direction="row" alignItems="center">
                            <Chip label={i18n(`newW.node.${step.type}`)} color="primary" size="small" />
                            <Typography component="div" variant="body1">
                              {step.name}
                            </Typography>
                          </Space>
                          <Space direction="row">
                            <IconButton
                              size="small"
                              title={i18n('common.edit', intl)}
                              onClick={restoreExperiment(index)}
                            >
                              <UndoIcon />
                            </IconButton>
                            <Menu>
                              <MenuItem dense onClick={handleSelect(step.name, index, 'delete')}>
                                <ListItemIcon>
                                  <RemoveIcon fontSize="small" />
                                </ListItemIcon>
                                <Typography variant="inherit">{i18n('common.delete')}</Typography>
                              </MenuItem>
                            </Menu>
                          </Space>
                        </Box>
                      </Paper>
                    </StepLabel>
                  ) : (
                    <Add externalTemplate={step} update={index} updateCallback={updateTemplateCallback} />
                  )}
                </Step>
              ))}
            {restoreIndex < 0 && (
              <Step>
                <Add />
              </Step>
            )}
          </Stepper>
        </Space>
      </Grid>
      <Grid item xs={12} md={4} className={classes.leftSticky}>
        <Formik
          initialValues={{ name: '', namespace: '', deadline: '' }}
          onSubmit={submitWorkflow}
          validate={onValidate}
          validateOnBlur={false}
        >
          {({ errors, touched }) => (
            <Form style={{ height: '100%' }}>
              <Space height="100%">
                <Typography>{i18n('newW.titleBasic')}</Typography>
                <TextField
                  name="name"
                  label={i18n('common.name')}
                  validate={validateName(i18n('newW.nameValidation', intl))}
                  helperText={errors.name && touched.name ? errors.name : i18n('newW.nameHelper')}
                  error={errors.name && touched.name ? true : false}
                />
                <SelectField
                  name="namespace"
                  label={i18n('k8s.namespace')}
                  helperText={i18n('newE.basic.namespaceHelper')}
                >
                  {namespaces!.map((n) => (
                    <MenuItem key={n} value={n}>
                      {n}
                    </MenuItem>
                  ))}
                </SelectField>
                <TextField
                  name="deadline"
                  label={i18n('newW.node.deadline')}
                  validate={validateDeadline(i18n('newW.node.deadlineValidation', intl))}
                  helperText={errors.deadline && touched.deadline ? errors.deadline : i18n('newW.node.deadlineHelper')}
                  error={errors.deadline && touched.deadline ? true : false}
                />
                <Typography>{i18n('common.preview')}</Typography>
                <Box flex={1}>
                  <Paper sx={{ p: 0 }}>
                    <YAMLEditor
                      data={constructWorkflow(workflowBasic, Object.values(templates))}
                      mountEditor={setYAMLEditor}
                    />
                  </Paper>
                </Box>
                <Button
                  type="submit"
                  variant="contained"
                  color="primary"
                  startIcon={<PublishIcon />}
                  fullWidth
                  disabled={_.isEmpty(templates)}
                >
                  {i18n('newW.submit')}
                </Button>
              </Space>
            </Form>
          )}
        </Formik>
      </Grid>
    </Grid>
  )
}

export default NewWorkflow
