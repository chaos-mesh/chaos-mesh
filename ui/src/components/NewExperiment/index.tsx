import { Box, Button, Container, Drawer } from '@material-ui/core'
import { Experiment, StepperFormProps } from './types'
import { Form, Formik, FormikHelpers } from 'formik'
import React, { useState } from 'react'
import { StepperProvider, useStepperContext } from './Context'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import PublishIcon from '@material-ui/icons/Publish'
import Stepper from './Stepper'
import api from 'api'
import { defaultExperimentSchema } from './constants'
import { parseSubmitValues } from 'lib/utils'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      display: 'flex',
      flexDirection: 'column',
      width: '50vw',
      height: '100%',
      padding: theme.spacing(6),
      [theme.breakpoints.down('sm')]: {
        width: '100vw',
      },
    },
  })
)

interface ActionsProps {
  isSubmitting?: boolean
}

const Actions = ({ isSubmitting = false }: ActionsProps) => {
  const { state } = useStepperContext()

  return (
    <Box display="flex" justifyContent="space-between" marginBottom={6}>
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

export default function NewExperiment() {
  const classes = useStyles()
  const initialValues: Experiment = defaultExperimentSchema

  const [isOpen, setIsOpen] = useState(false)
  const toggleDrawer = () => setIsOpen(!isOpen)

  const handleOnSubmit = (values: Experiment, actions: FormikHelpers<Experiment>) => {
    const parsedValues = parseSubmitValues(values)

    console.log(parsedValues)

    api.experiments
      .newExperiment(parsedValues)
      .then((resp) => {
        console.log(resp.data)
      })
      .catch(console.log)
  }

  return (
    <>
      <Button variant="outlined" startIcon={<AddIcon />} onClick={toggleDrawer}>
        New Experiment
      </Button>
      <Drawer anchor="right" open={isOpen} onClose={toggleDrawer}>
        <StepperProvider>
          <Formik initialValues={initialValues} onSubmit={handleOnSubmit}>
            {(props: StepperFormProps) => {
              const { isSubmitting } = props

              return (
                <Container className={classes.container}>
                  <Form>
                    <Actions isSubmitting={isSubmitting} />
                    <Stepper formProps={props} toggleDrawer={toggleDrawer} />
                  </Form>
                </Container>
              )
            }}
          </Formik>
        </StepperProvider>
      </Drawer>
    </>
  )
}
