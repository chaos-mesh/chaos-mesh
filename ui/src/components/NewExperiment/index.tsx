import { Box, Button, Container, Drawer, Fab, useMediaQuery, useTheme } from '@material-ui/core'
import { Form, Formik, FormikHelpers, useFormikContext } from 'formik'
import React, { useState } from 'react'
import { StepperProvider, useStepperContext } from './Context'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { defaultExperimentSchema, validationSchema } from './constants'
import { parseSubmitValues, yamlToExperiments } from 'lib/formikhelpers'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import AddIcon from '@material-ui/icons/Add'
import CancelIcon from '@material-ui/icons/Cancel'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import { Experiment } from './types'
import PublishIcon from '@material-ui/icons/Publish'
import Stepper from './Stepper'
import api from 'api'
import { setNeedToRefreshExperiments } from 'slices/experiments'
import { useHistory } from 'react-router-dom'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      height: '100%',
      padding: theme.spacing(6),
    },
    new: {
      [theme.breakpoints.down('xs')]: {
        display: 'none',
      },
    },
    fab: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      [theme.breakpoints.down('xs')]: {
        position: 'fixed',
        bottom: theme.spacing(7.5),
        right: theme.spacing(3),
        display: 'flex',
        background: theme.palette.primary.main,
        color: '#fff',
        zIndex: 1101, // .MuiAppBar-root z-index: 1100
      },
    },
    cancel: {
      zIndex: 1,
    },
  })
)

interface ActionsProps {
  toggleDrawer: () => void
  setInitialValues: (initialValues: Experiment) => void
}

const Actions = ({ toggleDrawer, setInitialValues }: ActionsProps) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const size = isTabletScreen ? ('small' as 'small') : ('medium' as 'medium')
  const classes = useStyles()

  const { isSubmitting, resetForm } = useFormikContext()

  const dispatch = useStoreDispatch()

  const { state } = useStepperContext()

  const handleCancel = () => {
    toggleDrawer()
    resetForm()
    setInitialValues(defaultExperimentSchema)
  }

  const handleUploadYAML = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files![0]

    const reader = new FileReader()
    reader.onload = function (e) {
      try {
        const y = yamlToExperiments(yaml.safeLoad(e.target!.result as string))
        setInitialValues(y)
        dispatch(
          setAlert({
            type: 'success',
            message: `Imported successfully!`,
          })
        )
      } catch (e) {
        console.error(e)
        dispatch(
          setAlert({
            type: 'error',
            message: `An error occurred: ${e}`,
          })
        )
      } finally {
        dispatch(setAlertOpen(true))
      }
    }
    reader.readAsText(f)
  }

  return (
    <Box display="flex" justifyContent="space-between" marginBottom={6}>
      <Button
        className={classes.cancel}
        variant="outlined"
        size={size}
        startIcon={<CancelIcon />}
        onClick={handleCancel}
      >
        Cancel
      </Button>
      <Box display="flex">
        <Box mr={3}>
          <Button variant="outlined" component="label" size={size} startIcon={<CloudUploadOutlinedIcon />}>
            Yaml File
            <input type="file" style={{ display: 'none' }} onChange={handleUploadYAML} />
          </Button>
        </Box>
        <Button
          type="submit"
          variant="contained"
          size={size}
          color="primary"
          startIcon={<PublishIcon />}
          disabled={state.activeStep < 4 || isSubmitting}
        >
          Submit
        </Button>
      </Box>
    </Box>
  )
}

export default function NewExperiment() {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const classes = useStyles()
  const history = useHistory()
  const dispatch = useStoreDispatch()

  const [initialValues, setInitialValues] = useState<Experiment>(defaultExperimentSchema)
  const [open, setOpen] = useState(false)
  const toggleDrawer = () => setOpen(!open)

  const handleOnSubmit = (values: Experiment, actions: FormikHelpers<Experiment>) => {
    const parsedValues = parseSubmitValues(values)

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug parsedValues:', parsedValues)
    }

    api.experiments
      .newExperiment(parsedValues)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: 'Created successfully!',
          })
        )
        dispatch(setAlertOpen(true))

        if (history.location.pathname === '/experiments') {
          dispatch(setNeedToRefreshExperiments(true))
        } else {
          history.push('/experiments')
        }
      })
      .catch(console.log)
      .finally(toggleDrawer)
  }

  return (
    <>
      <Button className={classes.new} variant="outlined" size="small" startIcon={<AddIcon />} onClick={toggleDrawer}>
        New Experiment
      </Button>
      <Fab className={classes.fab} color="inherit" size="medium" aria-label="New experiment">
        <AddIcon onClick={toggleDrawer} />
      </Fab>
      <Drawer
        anchor="right"
        open={open}
        onClose={toggleDrawer}
        PaperProps={{ style: { width: isTabletScreen ? '100vw' : '50vw' } }}
      >
        <StepperProvider>
          <Formik
            enableReinitialize
            initialValues={initialValues}
            validationSchema={validationSchema}
            onSubmit={handleOnSubmit}
          >
            <Container className={classes.container}>
              <Form>
                <Actions toggleDrawer={toggleDrawer} setInitialValues={setInitialValues} />
                <Stepper />
              </Form>
            </Container>
          </Formik>
        </StepperProvider>
      </Drawer>
    </>
  )
}
