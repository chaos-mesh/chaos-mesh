import { Box, Button, Step, StepLabel, Stepper, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { back, jump, next, reset, useStepperContext } from '../Context'

import BasicStep from './Basic'
import DoneAllIcon from '@material-ui/icons/DoneAll'
import Loading from 'components/Loading'
import ScheduleStep from './Schedule'
import ScopeStep from './Scope'
import { StepperFormProps } from '../types'
import TargetStep from './Target'
import api from 'api'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    main: {
      display: 'flex',
      flexDirection: 'column',
      width: '75%',
      margin: `0 auto`,
      padding: `${theme.spacing(6)} 0`,
      [theme.breakpoints.down('sm')]: {
        width: '100%',
      },
    },
    stepper: {
      [theme.breakpoints.down('sm')]: {
        paddingLeft: 0,
        paddingRight: 0,
      },
    },
    stepLabel: {
      cursor: 'pointer !important',
    },
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

  const { state, dispatch } = useStepperContext()
  const { activeStep } = state
  const [namespaces, setNamespaces] = useState<string[]>([])

  const fetchNamespaces = () => {
    api.common
      .namespaces()
      .then((resp) => setNamespaces(resp.data))
      .catch(console.log)
  }

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
            <Typography variant="h6">All steps are completed.</Typography>
          </Box>
        )
      default:
        return <BasicStep formProps={formProps} namespaces={namespaces} />
    }
  }

  const handleNext = () => dispatch(next())
  const handleBack = () => dispatch(back())
  const handleJump = (step: number) => () => dispatch(jump(step))
  const handleReset = () => {
    dispatch(reset())

    const { dirty, handleReset: resetForm } = formProps

    if (dirty) {
      resetForm()
    }
  }

  return (
    <Box display="flex" flexDirection="column">
      <Stepper className={classes.stepper} activeStep={state.activeStep} alternativeLabel>
        {steps.map((label, index) => (
          <Step key={label}>
            <StepLabel className={classes.stepLabel} onClick={handleJump(index)}>
              {label}
            </StepLabel>
          </Step>
        ))}
      </Stepper>

      {namespaces.length > 0 && (
        <Box className={classes.main}>
          <Box>{getStepContent()}</Box>
          <Box marginTop={6} textAlign="right">
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
      )}

      {namespaces.length === 0 && <Loading />}
    </Box>
  )
}

export default CreateStepper
