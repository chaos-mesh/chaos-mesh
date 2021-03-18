import { Box, Grid, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import Add from './Add'
import CheckIcon from '@material-ui/icons/Check'
import { Experiment } from 'components/NewExperiment/types'
import NewExperiment from 'components/NewExperimentNext'
import Paper from 'components-mui/Paper'
import UndoIcon from '@material-ui/icons/Undo'
import { makeStyles } from '@material-ui/core/styles'
import { resetNewExperiment } from 'slices/experiments'

const useStyles = makeStyles((theme) => ({
  stepper: {
    paddingTop: 0,
    paddingRight: 0,
    paddingLeft: 0,
  },
  success: {
    color: theme.palette.success.main,
  },
  submittedStep: {
    borderColor: theme.palette.success.main,
  },
  asButton: {
    cursor: 'pointer',
  },
}))

type IStep = Experiment

const NewWorkflow = () => {
  const classes = useStyles()

  const { templates } = useStoreSelector((state) => state.workflows)
  const dispatch = useStoreDispatch()

  const [steps, setSteps] = useState<IStep[]>([])

  useEffect(() => {
    return () => {
      dispatch(resetNewExperiment())
    }
  }, [dispatch])

  useEffect(() => {
    if (templates.length) {
      setSteps(templates.map((d) => d.experiment))
    }
  }, [templates])

  const onSubmitCallback = () => {
    dispatch(resetNewExperiment())
  }

  return (
    <Grid container>
      <Grid item xs={12} md={9}>
        <Stepper className={classes.stepper} orientation="vertical">
          {steps.length > 0 &&
            steps.map((step) => (
              <Step key={step.name}>
                <StepLabel icon={<CheckIcon className={classes.success} />}>
                  <Paper className={classes.submittedStep}>
                    <Box display="flex" justifyContent="space-between" alignItems="center">
                      <Typography component="div">{step.name}</Typography>
                      <UndoIcon className={classes.asButton} />
                    </Box>
                    <NewExperiment loadFrom={false} />
                  </Paper>
                </StepLabel>
              </Step>
            ))}
          <Step>
            <Add onSubmitCallback={onSubmitCallback} />
          </Step>
        </Stepper>
      </Grid>
    </Grid>
  )
}

export default NewWorkflow
