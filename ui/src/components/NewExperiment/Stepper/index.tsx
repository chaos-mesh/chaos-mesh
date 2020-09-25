import { Box, Button, Step, StepLabel, Stepper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { back, jump, next, reset, useStepperContext } from '../Context'

import BasicStep from './Basic'
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import DoneAllIcon from '@material-ui/icons/DoneAll'
import DoneIcon from '@material-ui/icons/Done'
import PublishIcon from '@material-ui/icons/Publish'
import ScheduleStep from './Schedule'
import ScopeStep from './Scope'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import TargetStep from './Target'
import { getNamespaces } from 'slices/experiments'
import { useFormikContext } from 'formik'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    main: {
      display: 'flex',
      flexDirection: 'column',
      margin: '0 auto',
      padding: theme.spacing(6),
      width: '80%',
      [theme.breakpoints.down('sm')]: {
        width: '100%',
      },
    },
    stepper: {
      background: 'none',
      [theme.breakpoints.down('sm')]: {
        paddingLeft: 0,
        paddingRight: 0,
        '& > .MuiStep-horizontal': {
          '&:first-child': {
            paddingLeft: 0,
          },
          '&:last-child': {
            paddingRight: 0,
          },
        },
      },
    },
    stepLabel: {
      cursor: 'pointer !important',
    },
    marginRight6: {
      marginRight: theme.spacing(6),
    },
  })
)

const steps = ['basic', 'scope', 'target', 'schedule']

const CreateStepper: React.FC = () => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const size = isTabletScreen ? ('small' as 'small') : ('medium' as 'medium')
  const classes = useStyles()

  const { resetForm, isSubmitting } = useFormikContext()

  const { namespaces } = useSelector((state: RootState) => state.experiments)
  const storeDispatch = useStoreDispatch()

  const { state, dispatch } = useStepperContext()
  const { activeStep } = state

  useEffect(() => {
    storeDispatch(getNamespaces())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleNext = () => dispatch(next())
  const handleBack = () => dispatch(back())
  const handleJump = (step: number) => () => dispatch(jump(step))
  const handleReset = () => {
    dispatch(reset())
    resetForm()
  }

  const getStepContent = () => {
    switch (activeStep) {
      case 0:
        return <BasicStep namespaces={namespaces} />
      case 1:
        return <ScopeStep namespaces={namespaces} />
      case 2:
        return <TargetStep />
      case 3:
        return <ScheduleStep />
      case 4:
        return (
          <Box textAlign="center" my={6}>
            <DoneAllIcon fontSize="large" />
            <Typography variant="h6">{T('newE.complete')}</Typography>
          </Box>
        )
      default:
        return <BasicStep namespaces={namespaces} />
    }
  }

  return (
    <Box display="flex" flexDirection="column" mt={6}>
      <Stepper className={classes.stepper} activeStep={state.activeStep} alternativeLabel>
        {steps.map((label, index) => (
          <Step key={label}>
            <StepLabel className={classes.stepLabel} onClick={handleJump(index)}>
              {T(`newE.steps.${label}`)}
            </StepLabel>
          </Step>
        ))}
      </Stepper>

      <Box className={classes.main}>
        {namespaces.length > 0 ? (
          <>
            <Box>{getStepContent()}</Box>
            <Box mt={6} textAlign="right">
              {activeStep === steps.length ? (
                <Box>
                  <Button className={classes.marginRight6} size={size} onClick={handleReset}>
                    {T('common.reset')}
                  </Button>
                  <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                    size={size}
                    startIcon={<PublishIcon />}
                    disabled={activeStep < 4 || isSubmitting}
                  >
                    {T('common.submit')}
                  </Button>
                </Box>
              ) : (
                <>
                  {activeStep !== 0 && (
                    <Button
                      className={classes.marginRight6}
                      size={size}
                      startIcon={<ChevronLeftIcon />}
                      onClick={handleBack}
                    >
                      {T('common.back')}
                    </Button>
                  )}
                  <Button
                    variant="contained"
                    color="primary"
                    size={size}
                    endIcon={activeStep === steps.length - 1 ? <DoneIcon /> : <ChevronRightIcon />}
                    onClick={handleNext}
                  >
                    {activeStep === steps.length - 1 ? T('common.finish') : T('common.next')}
                  </Button>
                </>
              )}
            </Box>
          </>
        ) : (
          <SkeletonN n={6} />
        )}
      </Box>
    </Box>
  )
}

export default CreateStepper
