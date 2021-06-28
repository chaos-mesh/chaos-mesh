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
} from '@material-ui/core'
import { Form, Formik } from 'formik'
import { SelectField, TextField } from 'components/FormField'
import { Template, deleteTemplate, resetWorkflow } from 'slices/workflows'
import { setAlert, setConfirm } from 'slices/globalStatus'
import { useEffect, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import { Ace } from 'ace-builds'
import Add from './Add'
import CheckIcon from '@material-ui/icons/Check'
import Menu from 'components-mui/Menu'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import RemoveIcon from '@material-ui/icons/Remove'
import Space from 'components-mui/Space'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import YAMLEditor from 'components/YAMLEditor'
import _isEmpty from 'lodash.isempty'
import api from 'api'
import { constructWorkflow } from 'lib/formikhelpers'
import { makeStyles } from '@material-ui/styles'
import { resetNewExperiment } from 'slices/experiments'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import yaml from 'js-yaml'

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
  const history = useHistory()

  const state = useStoreSelector((state) => state)
  const { namespaces } = state.experiments
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

  useEffect(() => {
    return () => {
      dispatch(resetNewExperiment())
    }
  }, [dispatch])

  useEffect(() => {
    setSteps(_isEmpty(templates) ? [] : templates)
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
        message: T('confirm.success.delete', intl) as string,
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
            title: `${T('common.delete', intl)} ${name}`,
            description: T('newW.node.deleteDesc', intl) as string,
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
    const workflow = yamlEditor?.getValue()

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug workflow:', workflow)
    }

    api.workflows
      .newWorkflow(yaml.load(workflow!))
      .then(() => {
        dispatch(resetWorkflow())

        history.push('/workflows')
      })
      .catch(console.error)
  }

  return (
    <>
      <Grid container spacing={9}>
        <Grid item xs={12} md={8}>
          <Space spacing={6}>
            <Typography>{T('common.process')}</Typography>
            <Stepper orientation="vertical" sx={{ mt: -1, p: 0 }}>
              {steps.length > 0 &&
                steps.map((step, index) => (
                  <Step key={step.name}>
                    {restoreIndex !== index ? (
                      <StepLabel icon={<CheckIcon sx={{ color: 'success.main' }} />}>
                        <Paper sx={{ p: 3, borderColor: 'success.main' }}>
                          <Box display="flex" justifyContent="space-between">
                            <Space direction="row" alignItems="center">
                              <Chip label={T(`newW.node.${step.type}`)} color="primary" size="small" />
                              <Typography component="div" variant="body1">
                                {step.name}
                              </Typography>
                            </Space>
                            <Space direction="row">
                              <IconButton
                                size="small"
                                title={T('common.edit', intl)}
                                onClick={restoreExperiment(index)}
                              >
                                <UndoIcon />
                              </IconButton>
                              <Menu>
                                <MenuItem dense onClick={handleSelect(step.name, index, 'delete')}>
                                  <ListItemIcon>
                                    <RemoveIcon fontSize="small" />
                                  </ListItemIcon>
                                  <Typography variant="inherit">{T('common.delete')}</Typography>
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
                  <Typography>{T('newW.titleBasic')}</Typography>
                  <TextField
                    name="name"
                    label={T('common.name')}
                    validate={validateName(T('newW.nameValidation', intl))}
                    helperText={errors.name && touched.name ? errors.name : T('newW.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />
                  <SelectField name="namespace" label={T('k8s.namespace')} helperText={T('newE.basic.namespaceHelper')}>
                    {namespaces.map((n) => (
                      <MenuItem key={n} value={n}>
                        {n}
                      </MenuItem>
                    ))}
                  </SelectField>
                  <TextField
                    name="deadline"
                    label={T('newW.node.deadline')}
                    validate={validateDeadline(T('newW.node.deadlineValidation', intl))}
                    helperText={errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')}
                    error={errors.deadline && touched.deadline ? true : false}
                  />
                  <Typography>{T('common.preview')}</Typography>
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
                    disabled={_isEmpty(templates)}
                  >
                    {T('newW.submit')}
                  </Button>
                </Space>
              </Form>
            )}
          </Formik>
        </Grid>
      </Grid>
    </>
  )
}

export default NewWorkflow
