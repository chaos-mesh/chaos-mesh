import { Box, Button, Step, StepLabel, Stepper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { back, jump, next, reset, useStepperContext } from '../Context'
import { getAnnotations, getLabels, getNamespaces } from 'slices/experiments'

import BasicStep from './Basic'
import DoneAllIcon from '@material-ui/icons/DoneAll'
import Loading from 'components/Loading'
import ScheduleStep from './Schedule'
import ScopeStep from './Scope'
import { StepperFormProps } from '../types'
import TargetStep from './Target'
import { useSelector } from 'react-redux'

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
      '& > .MuiStep-horizontal': {
        '&:first-child': {
          paddingLeft: 0,
        },
        '&:last-child': {
          paddingRight: 0,
        },
      },
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
}

const CreateStepper: React.FC<StepperProps> = ({ formProps }) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const size = isTabletScreen ? ('small' as 'small') : ('medium' as 'medium')
  const classes = useStyles()

  const { namespaces, labels, annotations } = useSelector((state: RootState) => state.experiments)
  const storeDispatch = useStoreDispatch()

  const { state, dispatch } = useStepperContext()
  const { activeStep } = state

  useEffect(() => {
    storeDispatch(getNamespaces())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (namespaces.length > 0) {
      storeDispatch(getLabels(namespaces.join(',')))
      storeDispatch(getAnnotations(namespaces.join(',')))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [namespaces])

  const handleNext = () => dispatch(next())
  const handleBack = () => dispatch(back())
  const handleJump = (step: number) => () => dispatch(jump(step))
  const handleReset = () => {
    dispatch(reset())
    formProps.resetForm()
  }

  const getStepContent = () => {
    switch (activeStep) {
      case 0:
        return <BasicStep formProps={formProps} namespaces={namespaces} />
      case 1:
        return <ScopeStep formProps={formProps} namespaces={namespaces} labels={labels} annotations={annotations} />
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
          <Box mt={6} textAlign="right">
            {activeStep === steps.length ? (
              <Button size={size} onClick={handleReset}>
                Reset
              </Button>
            ) : (
              <>
                {activeStep !== 0 && (
                  <Button className={classes.backButton} size={size} onClick={handleBack}>
                    Back
                  </Button>
                )}
                <Button variant="contained" color="primary" size={size} onClick={handleNext}>
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
