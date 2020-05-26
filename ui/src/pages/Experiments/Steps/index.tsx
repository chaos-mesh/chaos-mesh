import React, { useState, useEffect } from 'react'
import { Box, Button, Stepper, Step, StepLabel, Typography } from '@material-ui/core'
import DoneAllIcon from '@material-ui/icons/DoneAll'
import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

import BasicStep from './Basic'
import ScopeStep from './Scope'
import TargetStep from './Target'
import ScheduleStep from './Schedule'
import { useStepperContext } from '../Context'
import { StepProps } from '../types'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    backButton: {
      marginRight: theme.spacing(5),
    },
  })
)

const steps = ['Basic', 'Scope', 'Target', 'Schedule']

export default function CreateStepper({ formProps }: StepProps) {
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
    const props = {
      formProps,
    }

    switch (activeStep) {
      case 0:
        return <BasicStep {...props} namespaces={namespaces} />
      case 1:
        return <ScopeStep {...props} namespaces={namespaces} />
      case 2:
        return <TargetStep {...props} />
      case 3:
        return <ScheduleStep {...props} />
      case 4:
        return (
          <Box textAlign="center">
            <DoneAllIcon fontSize="large" />
            <Typography variant="h6">All steps completed.</Typography>
          </Box>
        )
      default:
        return 'Unknown stepIndex'
    }
  }

  const handleNext = () => {
    dispatch({ type: 'next' })
  }

  const handleBack = () => {
    dispatch({ type: 'back' })
  }

  const handleReset = () => {
    dispatch({ type: 'reset' })

    const { dirty, handleReset: restForm } = formProps
    if (dirty) {
      restForm()
    }
  }

  return (
    <Box display="flex" flexDirection="column" flexGrow={1}>
      <Stepper activeStep={state.activeStep} alternativeLabel>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>

      <Box display="flex" flexDirection="column" flexGrow={1} py={4} px={{ xs: 8, sm: 24, md: 32, lg: 40, xl: 42 }}>
        <Box flexGrow={1} width="100%">
          {getStepContent()}
        </Box>

        <Box mt={4} textAlign="right">
          {activeStep === steps.length ? (
            <Button onClick={handleReset}>Reset</Button>
          ) : (
            <>
              <Button disabled={activeStep === 0} onClick={handleBack} className={classes.backButton}>
                Back
              </Button>
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
