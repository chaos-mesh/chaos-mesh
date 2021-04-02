import { Box, Button, Grid, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import { Form, Formik } from 'formik'
import MultiNode, { MultiNodeHandles } from './MultiNode'
import { Template, deleteTemplate, updateTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import Add from './Add'
import CheckIcon from '@material-ui/icons/Check'
import NewExperiment from 'components/NewExperimentNext'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline'
import Space from 'components-mui/Space'
import T from 'components/T'
import { TextField } from 'components/FormField'
import UndoIcon from '@material-ui/icons/Undo'
import YAMLEditor from 'components/YAMLEditor'
import _isEmpty from 'lodash.isempty'
import _snakecase from 'lodash.snakecase'
import clsx from 'clsx'
import { constructWorkflow } from 'lib/formikhelpers'
import { makeStyles } from '@material-ui/core/styles'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'

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

const NewWorkflow = () => {
  const classes = useStyles()
  const intl = useIntl()

  const state = useStoreSelector((state) => state)
  const { templates } = state.workflows
  const { theme } = state.settings
  const dispatch = useStoreDispatch()

  const [steps, setSteps] = useState<IStep[]>([])
  const [restoreIndex, setRestoreIndex] = useState(-1)
  const [showRemove, setShowRemove] = useState(-1)
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
      const e = experiments[0]

      const kind = e.target.kind

      dispatch(
        setExternalExperiment({
          kindAction: [kind, e.target[_snakecase(kind)].action ?? ''],
          target: e.target,
          basic: e.basic,
        })
      )

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

  // TODO
  const submitWorkflow = () => {}

  return (
    <>
      <Grid container spacing={6}>
        <Grid item xs={12} md={8}>
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
                            <Formik initialValues={{ name: step.name }} onSubmit={() => {}}>
                              <Form>
                                <Box display="flex" justifyContent="space-between" alignItems="center" mb={6}>
                                  <TextField
                                    mb={0}
                                    className={classes.field}
                                    name="name"
                                    label={T('newE.basic.name')}
                                  />
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
                                </Box>
                              </Form>
                            </Formik>
                          )}
                          <NewExperiment loadFrom={false} onSubmit={onRestoreSubmit(step.type, index)} />
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
        </Grid>
        <Grid item xs={12} md={4} className={classes.leftSticky}>
          <Space display="flex" flexDirection="column" height="100%" vertical spacing={6}>
            <Box display="flex" justifyContent="space-between">
              <Button
                variant="contained"
                color="primary"
                startIcon={<PublishIcon />}
                disabled={_isEmpty(templates)}
                onClick={submitWorkflow}
              >
                {T('newW.submit')}
              </Button>
            </Box>
            <Typography>{T('common.preview')}</Typography>
            <Box flex={1}>
              <Paper style={{ height: '100%' }} padding={0}>
                <YAMLEditor theme={theme} data={constructWorkflow('test', '120s', Object.values(templates))} />
              </Paper>
            </Box>
          </Space>
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
