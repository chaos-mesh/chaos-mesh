import { Box, Button, Grid, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import { Template, updateTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import Add from './Add'
import CheckIcon from '@material-ui/icons/Check'
import NewExperiment from 'components/NewExperimentNext'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import _isEmpty from 'lodash.isempty'
import _snakecase from 'lodash.snakecase'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  stepper: {
    paddingTop: 0,
    paddingRight: 0,
    paddingLeft: 0,
  },
  success: {
    color: theme.palette.success.main,
  },
  primary: {
    color: theme.palette.primary.main,
  },
  submittedStep: {
    borderColor: theme.palette.success.main,
  },
  asButton: {
    cursor: 'pointer',
  },
}))

type IStep = Template

const NewWorkflow = () => {
  const classes = useStyles()

  const { templates } = useStoreSelector((state) => state.workflows)
  const dispatch = useStoreDispatch()

  const [steps, setSteps] = useState<IStep[]>([])
  const [restoreIndex, setRestoreIndex] = useState(-1)

  useEffect(() => {
    return () => {
      dispatch(resetNewExperiment())
    }
  }, [dispatch])

  useEffect(() => {
    if (!_isEmpty(templates)) {
      setSteps(Object.values(templates).sort((a, b) => a.index! - b.index!))
    }
  }, [templates])

  const onSubmitCallback = () => dispatch(resetNewExperiment())

  const restoreExperiment = (e: any, index: number) => () => {
    if (restoreIndex !== -1) {
      dispatch(resetNewExperiment())
      setRestoreIndex(-1)
    } else {
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

  const onRestoreSubmit = (type: Template['type'], index: number) => (experiment: any) => {
    setRestoreIndex(-1)

    if (type === 'single') {
      dispatch(
        updateTemplate({
          type,
          index,
          name: experiment.basic.name,
          experiments: [experiment],
        })
      )
    }
  }

  const submitWorkflow = () => {}

  return (
    <Grid container>
      <Grid item xs={12} md={8}>
        <Stepper className={classes.stepper} orientation="vertical">
          {steps.length > 0 &&
            steps.map((step, index) => (
              <Step key={step.type + index}>
                <StepLabel icon={<CheckIcon className={classes.success} />}>
                  <Paper className={classes.submittedStep}>
                    <Box display="flex" justifyContent="space-between" alignItems="center">
                      <Typography component="div" variant={restoreIndex === index ? 'h6' : 'body1'}>
                        {step.name}
                      </Typography>
                      <UndoIcon
                        className={classes.asButton}
                        onClick={restoreExperiment(step.experiments, step.index!)}
                      />
                    </Box>
                    {restoreIndex === index && (
                      <Box mt={3}>
                        <NewExperiment loadFrom={false} onSubmit={onRestoreSubmit(step.type, index)} />
                      </Box>
                    )}
                  </Paper>
                </StepLabel>
              </Step>
            ))}
          <Step>
            <Add onSubmitCallback={onSubmitCallback} />
          </Step>
          {!_isEmpty(templates) && (
            <Step>
              <StepLabel icon={<PublishIcon className={classes.primary} />}>
                <Button variant="contained" color="primary" fullWidth onClick={submitWorkflow}>
                  {T('newW.submit')}
                </Button>
              </StepLabel>
            </Step>
          )}
        </Stepper>
      </Grid>
    </Grid>
  )
}

export default NewWorkflow
