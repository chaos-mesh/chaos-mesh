import { Box, Button, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { back, next, reset, useStepperContext } from '../Context'

import BasicStep from './Basic'
import DoneAllIcon from '@material-ui/icons/DoneAll'
import ScheduleStep from './Schedule'
import ScopeStep from './Scope'
import { StepperFormProps } from '../types'
import TargetStep from './Target'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    backButton: {
      marginRight: theme.spacing(6),
    },
  })
)

const steps = ['Basic', 'Scope', 'Target', 'Schedule']

interface StepperProps {
  formProps: StepperFormProps
  toggleDrawer: () => void
}

const CreateStepper: React.FC<StepperProps> = ({ formProps, toggleDrawer }) => {
  const classes = useStyles()

  const [namespaces, setNamespaces] = useState<string[]>([])
  const { state, dispatch } = useStepperContext()
  const { activeStep } = state

  const fetchNamespaces = () => {
    // TODO: move mock data
    setNamespaces(['Default', 'Chaos Testing', 'Others'])
  }

  // fetch namespaces for basic and scope step when drawer is open
  useEffect(fetchNamespaces, [])

  const getStepContent = () => {
    switch (activeStep) {
      case 0:
        return <BasicStep formProps={formProps} namespaces={namespaces} />
      case 1:
        return <ScopeStep formProps={formProps} namespaces={namespaces} />
      case 2:
        return <TargetStep formProps={formProps} />
      case 3:
        return <ScheduleStep formProps={formProps} />
      case 4:
        return (
          <Box textAlign="center">
            <DoneAllIcon fontSize="large" />
            <Typography variant="h6">All steps completed.</Typography>
          </Box>
        )
      default:
        return <BasicStep formProps={formProps} namespaces={namespaces} />
    }
  }

  const handleNext = () => dispatch(next())

  const handleBack = () => dispatch(back())

  const handleReset = () => {
    dispatch(reset())

    const { dirty, handleReset: resetForm } = formProps

    if (dirty) {
      resetForm()
    }
  }

  return (
    <Box display="flex" flexDirection="column">
      <Stepper activeStep={state.activeStep} alternativeLabel>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>

      <Box display="flex" flexDirection="column" paddingY={6}>
        <Box>{getStepContent()}</Box>

        <Box mt={6} textAlign="right">
          {activeStep === steps.length ? (
            <Button onClick={handleReset}>Reset</Button>
          ) : (
            <>
              {activeStep === 0 ? (
                <Button className={classes.backButton} onClick={toggleDrawer}>
                  Cancel
                </Button>
              ) : (
                <Button className={classes.backButton} onClick={handleBack}>
                  Back
                </Button>
              )}
              <Button variant="contained" color="primary" onClick={handleNext}>
                {activeStep === steps.length - 1 ? 'Finish' : 'Next'}
              </Button>
            </>
          )}
        </Box>
      </Box>
    </Box>
  )
}

export default CreateStepper
