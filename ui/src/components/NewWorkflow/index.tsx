import { Box, Button, Grid, MenuItem, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import { Form, Formik } from 'formik'
import MultiNode, { MultiNodeHandles } from './MultiNode'
import { SelectField, TextField } from 'components/FormField'
import Suspend, { SuspendValues } from './Suspend'
import { Template, deleteTemplate, updateTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'
import { validateDuration, validateName } from 'lib/formikhelpers'

import { Ace } from 'ace-builds'
import Add from './Add'
import CheckIcon from '@material-ui/icons/Check'
import NewExperiment from 'components/NewExperimentNext'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline'
import Space from 'components-mui/Space'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import YAMLEditor from 'components/YAMLEditor'
import _isEmpty from 'lodash.isempty'
import _snakecase from 'lodash.snakecase'
import api from 'api'
import clsx from 'clsx'
import { constructWorkflow } from 'lib/formikhelpers'
import { makeStyles } from '@material-ui/core/styles'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import yaml from 'js-yaml'

const initialSelected = {
  name: '',
  title: '',
  description: '',
  action: '',
}

const useStyles = makeStyles((theme) => ({
  stepper: {
    padding: 0,
  },
  primary: {
    color: theme.palette.primary.main,
  },
  success: {
    color: theme.palette.success.main,
  },
  error: {
    color: theme.palette.error.main,
  },
  submittedStep: {
    borderColor: theme.palette.success.main,
  },
  removeSubmittedStep: {
    borderColor: theme.palette.error.main,
  },
  asButton: {
    cursor: 'pointer',
  },
  leftSticky: {
    position: 'sticky',
    top: 0,
    height: `calc(100vh - 56px - ${theme.spacing(12)})`,
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
  duration: string
}

const NewWorkflow = () => {
  const classes = useStyles()
  const intl = useIntl()

  const state = useStoreSelector((state) => state)
  const { namespaces } = state.experiments
  const { templates } = state.workflows
  const { theme } = state.settings
  const dispatch = useStoreDispatch()

  const [steps, setSteps] = useState<IStep[]>([])
  const [restoreIndex, setRestoreIndex] = useState(-1)
  const [showRemove, setShowRemove] = useState(-1)
  const [workflowBasic, setWorkflowBasic] = useState<WorkflowBasic>({
    name: '',
    namespace: '',
    duration: '',
  })
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()
  const [selected, setSelected] = useState(initialSelected)
  const confirmRef = useRef<ConfirmDialogHandles>(null)
  const multiNodeRef = useRef<MultiNodeHandles>(null)

  useEffect(() => {
    return () => {
      dispatch(resetNewExperiment())
    }
  }, [dispatch])

  useEffect(() => {
    if (!_isEmpty(templates)) {
      setSteps(Object.values(templates).sort((a, b) => a.index! - b.index!))
    } else {
      setSteps([])
    }
  }, [templates])

  const resetRestore = () => {
    dispatch(resetNewExperiment())
    setRestoreIndex(-1)
  }

  const restoreExperiment = (experiments: any, index: number) => () => {
    if (restoreIndex !== -1) {
      resetRestore()
    } else {
      if (experiments.length) {
        const e = experiments[0]

        const kind = e.target.kind

        dispatch(
          setExternalExperiment({
            kindAction: [kind, e.target[_snakecase(kind)].action ?? ''],
            target: e.target,
            basic: e.basic,
          })
        )
      } else {
      }

      setRestoreIndex(index)
    }
  }

  const setCurrentCallback = (experiments: Template['experiments']) => (index: number) => {
    const e = experiments[index]

    const kind = e.target.kind

    dispatch(
      setExternalExperiment({
        kindAction: [kind, e.target[_snakecase(kind)].action ?? ''],
        target: e.target,
        basic: e.basic,
      })
    )

    return true
  }

  const onRestoreSubmit = (type: Template['type'], index: number) => (experiment: any) => {
    if (type === 'single') {
      dispatch(
        updateTemplate({
          type,
          index,
          name: experiment.basic.name,
          experiments: [experiment],
        })
      )
      dispatch(
        setAlert({
          type: 'success',
          message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
        })
      )
      resetRestore()
    } else if (type === 'serial' || type === 'parallel') {
      const eIndex = multiNodeRef.current!.current
      const tmpSteps = [...steps]
      const tmpStep = { ...tmpSteps[index] }
      const tmpStepExperiments = tmpStep.experiments.slice()

      tmpStepExperiments[eIndex] = experiment
      tmpStep.experiments = tmpStepExperiments
      tmpSteps[index] = tmpStep

      setSteps(tmpSteps)

      dispatch(resetNewExperiment())

      if (eIndex !== tmpStep.experiments.length - 1) {
        setCurrentCallback(tmpStepExperiments)(eIndex + 1)
      }

      multiNodeRef.current!.setCurrent(eIndex + 1)
    }
  }

  const onNoSingleRestoreSubmit = (stepIndex: number) => () => {
    dispatch(updateTemplate(steps[stepIndex]))
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
      })
    )
    resetRestore()
  }

  const onSuspendRestoreSubmit = (stepIndex: number) => ({ name, duration }: SuspendValues) => {
    dispatch(
      updateTemplate({
        ...steps[stepIndex],
        name,
        duration,
      })
    )
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
      })
    )
    resetRestore()
  }

  const removeExperiment = (name: string) => {
    dispatch(deleteTemplate(name))
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.deleteSuccessfully' }),
      })
    )
    resetRestore()
  }

  const handleSelect = (name: string, action: string) => () => {
    switch (action) {
      case 'delete':
        setSelected({
          name,
          title: `${intl.formatMessage({ id: 'common.delete' })} ${name}`,
          description: intl.formatMessage({ id: 'newW.node.deleteDesc' }),
          action,
        })
        break
    }

    confirmRef.current!.setOpen(true)
  }

  const handleAction = (action: string) => () => {
    switch (action) {
      case 'delete':
        removeExperiment(selected.name)
        break
    }

    confirmRef.current!.setOpen(false)
  }

  const onValidate = setWorkflowBasic

  // TODO
  const submitWorkflow = () => {
    const workflow = yamlEditor?.getValue()

    console.log(yaml.load(workflow!))
    api.workflows.newWorkflow(yaml.load(workflow!))
  }

  return (
    <>
      <Grid container spacing={6}>
        <Grid item xs={12} md={8}>
          <Space vertical spacing={6}>
            <Typography>{T('common.process')}</Typography>
            <Stepper className={classes.stepper} orientation="vertical">
              {steps.length > 0 &&
                steps.map((step, index) => (
                  <Step key={step.type + index}>
                    <StepLabel
                      icon={
                        <Box
                          position="relative"
                          display="flex"
                          alignItems="center"
                          onMouseEnter={() => setShowRemove(index)}
                          onMouseLeave={() => setShowRemove(-1)}
                        >
                          <CheckIcon
                            className={classes.success}
                            style={{ visibility: showRemove === index ? 'hidden' : 'unset' }}
                          />
                          {showRemove === index && (
                            <Box position="absolute" top={0} title={intl.formatMessage({ id: 'common.delete' })}>
                              <RemoveCircleOutlineIcon
                                className={clsx(classes.error, classes.asButton)}
                                onClick={handleSelect(step.name, 'delete')}
                              />
                            </Box>
                          )}
                        </Box>
                      }
                    >
                      <Paper className={showRemove === index ? classes.removeSubmittedStep : classes.submittedStep}>
                        <Box display="flex" justifyContent="space-between">
                          <Typography component="div" variant={restoreIndex === index ? 'h6' : 'body1'}>
                            {step.name}
                          </Typography>
                          <UndoIcon
                            className={classes.asButton}
                            onClick={restoreExperiment(step.experiments, step.index!)}
                          />
                        </Box>
                        {restoreIndex === index && (
                          <Box mt={6}>
                            {(step.type === 'serial' || step.type === 'parallel') && (
                              <Formik initialValues={{ name: step.name, duration: step.duration }} onSubmit={() => {}}>
                                <Form>
                                  <Box display="flex" justifyContent="space-between" alignItems="center" mb={6}>
                                    <Space>
                                      <TextField className={classes.field} name="name" label={T('newE.basic.name')} />
                                      <TextField
                                        className={classes.field}
                                        name="duration"
                                        label={T('newE.schedule.duration')}
                                      />
                                    </Space>
                                    <Space display="flex">
                                      <MultiNode
                                        ref={multiNodeRef}
                                        count={step.experiments.length}
                                        setCurrentCallback={setCurrentCallback(step.experiments)}
                                      />
                                      <Button
                                        variant="contained"
                                        color="primary"
                                        startIcon={<PublishIcon />}
                                        onClick={onNoSingleRestoreSubmit(index)}
                                      >
                                        {T('newW.node.submitAll')}
                                      </Button>
                                    </Space>
                                  </Box>
                                </Form>
                              </Formik>
                            )}
                            {step.type !== 'suspend' && (
                              <NewExperiment loadFrom={false} onSubmit={onRestoreSubmit(step.type, index)} />
                            )}
                            {step.type === 'suspend' && (
                              <Suspend
                                initialValues={{
                                  name: steps[index].name,
                                  duration: steps[index].duration!,
                                }}
                                onSubmit={onSuspendRestoreSubmit(index)}
                              />
                            )}
                          </Box>
                        )}
                      </Paper>
                    </StepLabel>
                  </Step>
                ))}
              <Step>
                <Add />
              </Step>
            </Stepper>
          </Space>
        </Grid>
        <Grid item xs={12} md={4} className={classes.leftSticky}>
          <Formik
            initialValues={{ name: '', namespace: '', duration: '' }}
            onSubmit={() => {}}
            validate={onValidate}
            validateOnBlur={false}
          >
            {({ errors, touched }) => (
              <Space display="flex" flexDirection="column" height="100%" vertical spacing={6}>
                <Typography>{T('common.preview')}</Typography>
                <Form>
                  <TextField
                    name="name"
                    label={T('newE.basic.name')}
                    validate={validateName((T('newW.nameValidation') as unknown) as string)}
                    helperText={errors.name && touched.name ? errors.name : T('newW.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />
                  <SelectField
                    name="namespace"
                    label={T('newE.basic.namespace')}
                    helperText={T('newE.basic.namespaceHelper')}
                  >
                    {namespaces.map((n) => (
                      <MenuItem key={n} value={n}>
                        {n}
                      </MenuItem>
                    ))}
                  </SelectField>
                  <TextField
                    name="duration"
                    label={T('newE.schedule.duration')}
                    validate={validateDuration((T('newW.durationValidation') as unknown) as string)}
                    helperText={errors.duration && touched.duration ? errors.duration : T('newW.durationHelper')}
                    error={errors.duration && touched.duration ? true : false}
                  />
                </Form>
                <Box flex={1}>
                  <Paper style={{ height: '100%' }} padding={0}>
                    <YAMLEditor
                      theme={theme}
                      data={constructWorkflow(workflowBasic, Object.values(templates))}
                      mountEditor={setYAMLEditor}
                    />
                  </Paper>
                </Box>
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={<PublishIcon />}
                  fullWidth
                  disabled={_isEmpty(templates)}
                  onClick={submitWorkflow}
                >
                  {T('newW.submit')}
                </Button>
              </Space>
            )}
          </Formik>
        </Grid>
      </Grid>
      <ConfirmDialog
        ref={confirmRef}
        title={selected.title}
        description={selected.description}
        onConfirm={handleAction(selected.action)}
      />
    </>
  )
}

export default NewWorkflow
