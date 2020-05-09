import React from 'react'
import { Box, Button, Container, Stepper, Step, StepLabel, Typography } from '@material-ui/core'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import PublishIcon from '@material-ui/icons/Publish'

import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      display: 'flex',
      flexDirection: 'column',
      width: '80vw',
      height: '100%',
      padding: theme.spacing(6),
    },
    stepContent: {
      maxWidth: '80%',
      margin: 'auto',
      padding: theme.spacing(2),
    },
    stepActions: {
      marginTop: theme.spacing(4),
      textAlign: 'right',
    },
    backButton: {
      marginRight: theme.spacing(5),
    },
    resetButton: { marginTop: theme.spacing(4) },
    instructions: {
      textAlign: 'center',
    },
    finalTip: {
      textAlign: 'center',
    },
  })
)

function getSteps() {
  return ['Basic', 'Scope', 'Target', 'Schedule']
}

// TODO: mock demo, use form fields
function getStepContent(stepIndex: number) {
  switch (stepIndex) {
    case 0:
      return 'basic settings...'
    case 1:
      return 'scope settings...'
    case 2:
      return 'target settings...'
    case 3:
      return 'schedule settings...'
    default:
      return 'Unknown stepIndex'
  }
}

const CreateStepper = () => {
  const classes = useStyles()
  const [activeStep, setActiveStep] = React.useState(0)
  const steps = getSteps()

  const handleNext = () => {
    setActiveStep((prevActiveStep) => prevActiveStep + 1)
  }

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1)
  }

  const handleReset = () => {
    setActiveStep(0)
  }

  return (
    <Box flexGrow={1}>
      <Stepper activeStep={activeStep} alternativeLabel>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>
      <div className={classes.stepContent}>
        {activeStep === steps.length ? (
          <div className={classes.finalTip}>
            <Typography className={classes.instructions}>All steps completed.</Typography>
            <Button className={classes.resetButton} onClick={handleReset}>
              Reset
            </Button>
          </div>
        ) : (
          <>
            <Typography className={classes.instructions}>{getStepContent(activeStep)}</Typography>
            <div className={classes.stepActions}>
              <Button disabled={activeStep === 0} onClick={handleBack} className={classes.backButton}>
                Back
              </Button>
              <Button variant="contained" color="primary" onClick={handleNext}>
                {activeStep === steps.length - 1 ? 'Finish' : 'Next'}
              </Button>
            </div>
          </>
        )}
      </div>
    </Box>
  )
}

export default function NewExperiment() {
  const classes = useStyles()

  return (
    <Container className={classes.container} maxWidth="lg">
      <Box display="flex" justifyContent="space-between" mb={6}>
        <Button variant="outlined" startIcon={<CloudUploadOutlinedIcon />}>
          Yaml File
        </Button>
        <Button variant="contained" color="primary" startIcon={<PublishIcon />}>
          Submit
        </Button>
      </Box>

      <CreateStepper />
    </Container>
  )
}
