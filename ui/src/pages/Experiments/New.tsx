import React from 'react'
import { Box, Button, Container } from '@material-ui/core'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import PublishIcon from '@material-ui/icons/Publish'
import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'
import { Formik, Form, FormikHelpers } from 'formik'

import CreateStepper from './Steps'
import { Experiment, StepperFormProps } from './types'
import { StepperProvider, useStepperContext } from './Context'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    form: {
      height: '100%',
    },
    container: {
      display: 'flex',
      flexDirection: 'column',
      width: '80vw',
      height: '100%',
      padding: theme.spacing(6),
    },
  })
)

interface FormActionsProps {
  isSubmitting?: boolean
}

const FormActions = ({ isSubmitting = false }: FormActionsProps) => {
  const { state } = useStepperContext()

  return (
    <Box display="flex" justifyContent="space-between" mb={6}>
      <Button type="button" variant="outlined" startIcon={<CloudUploadOutlinedIcon />}>
        Yaml File
      </Button>
      <Button
        type="submit"
        variant="contained"
        color="primary"
        startIcon={<PublishIcon />}
        disabled={state.activeStep < 4 || isSubmitting}
      >
        Submit
      </Button>
    </Box>
  )
}

const defaultExperiment = {
  basic: {
    name: '',
    namespace: '',
  },
  scope: {
    namespaceSelector: [],
    phaseSelector: [],
    mode: 'all',
    value: '',
  },
  target: {
    pod: {
      action: '',
      container: '',
    },
    network: {
      action: '',
      delay: {
        latency: '',
        correlation: '',
        jitter: '',
      },
    },
  },
  schedule: {
    cron: '',
    duration: '',
  },
}

export default function NewExperiment() {
  const classes = useStyles()
  const initialValues: Experiment = defaultExperiment

  const handleSubmit = (values: Experiment, formikHelpers: FormikHelpers<Experiment>) => {
    console.log({ values, formikHelpers })
  }

  return (
    <StepperProvider>
      {/* Formik:Build forms in React, without the tears. */}
      {/* https://github.com/jaredpalmer/formik */}
      <Formik initialValues={initialValues} onSubmit={handleSubmit}>
        {(props: StepperFormProps) => {
          const { isSubmitting } = props

          return (
            <Form autoComplete="off" className={classes.form}>
              <Container className={classes.container} maxWidth="lg">
                <FormActions isSubmitting={isSubmitting} />
                <CreateStepper formProps={props} />
              </Container>
            </Form>
          )
        }}
      </Formik>
    </StepperProvider>
  )
}
