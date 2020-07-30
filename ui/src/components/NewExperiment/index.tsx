import { Box, Button, Divider, FormControlLabel, Grid, Paper, Radio, RadioGroup, Typography } from '@material-ui/core'
import { Form, Formik, FormikHelpers } from 'formik'
import React, { useEffect, useState } from 'react'
import { defaultExperimentSchema, validationSchema } from './constants'
import { parseLoaded, parseSubmit, yamlToExperiment } from 'lib/formikhelpers'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import { Archive } from 'api/archives.type'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import { Experiment } from './types'
import { Experiment as ExperimentResponse } from 'api/experiments.type'
import PaperTop from 'components/PaperTop'
import SkeletonN from 'components/SkeletonN'
import Stepper from './Stepper'
import { StepperProvider } from './Context'
import api from 'api'
import { setNeedToRefreshExperiments } from 'slices/experiments'
import { useHistory } from 'react-router-dom'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const Skeleton3 = () => <SkeletonN n={3} />

const LoadWrapper: React.FC<{ title: string }> = ({ title, children }) => (
  <Box mb={6}>
    <Box mb={6}>
      <Typography>{title}</Typography>
    </Box>
    {children}
  </Box>
)

interface ActionsProps {
  setInitialValues: (initialValues: Experiment) => void
}

const Actions = ({ setInitialValues }: ActionsProps) => {
  const dispatch = useStoreDispatch()

  const [experiments, setExperiments] = useState<ExperimentResponse[] | null>(null)
  const [archives, setArchives] = useState<Archive[] | null>(null)
  const [experimentRadio, setExperimentRadio] = useState('')
  const [archiveRadio, setArchiveRadio] = useState('')

  const onExperimentRadioChange = (e: any) => {
    const uuid = e.target.value

    setExperimentRadio(uuid)
    setArchiveRadio('')

    api.experiments
      .detail(uuid)
      .then(({ data }) => setInitialValues(parseLoaded(data.experiment_info)))
      .catch(console.log)
  }

  const onArchiveRadioChange = (e: any) => {
    const uuid = e.target.value

    setArchiveRadio(uuid)
    setExperimentRadio('')

    api.archives
      .detail(uuid)
      .then(({ data }) => setInitialValues(parseLoaded(data.experiment_info)))
      .catch(console.log)
  }

  const fetchExperiments = () =>
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.log)

  const fetchArchives = () =>
    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.log)

  useEffect(() => {
    fetchExperiments()
    fetchArchives()
  }, [])

  const handleUploadYAML = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files![0]

    const reader = new FileReader()
    reader.onload = function (e) {
      try {
        const y = yamlToExperiment(yaml.safeLoad(e.target!.result as string))
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
    <Box p={6}>
      <LoadWrapper title="Load From Existing Experiments">
        <RadioGroup value={experimentRadio} onChange={onExperimentRadioChange}>
          {experiments && experiments.length > 0 ? (
            experiments.map((e) => (
              <FormControlLabel key={e.uid} value={e.uid} control={<Radio color="primary" />} label={e.name} />
            ))
          ) : experiments?.length === 0 ? (
            <Typography variant="body2">No experiments found.</Typography>
          ) : (
            <Skeleton3 />
          )}
        </RadioGroup>
      </LoadWrapper>

      <Box my={6}>
        <Divider />
      </Box>

      <LoadWrapper title="Load From Archives">
        <RadioGroup value={archiveRadio} onChange={onArchiveRadioChange}>
          {archives && archives.length > 0 ? (
            archives.map((a) => (
              <FormControlLabel key={a.uid} value={a.uid} control={<Radio color="primary" />} label={a.name} />
            ))
          ) : archives?.length === 0 ? (
            <Typography variant="body2">No archives found.</Typography>
          ) : (
            <Skeleton3 />
          )}
        </RadioGroup>
      </LoadWrapper>

      <Box my={6}>
        <Divider />
      </Box>

      <LoadWrapper title="Load From YAML File">
        <Button component="label" variant="outlined" size="small" startIcon={<CloudUploadOutlinedIcon />}>
          Upload
          <input type="file" style={{ display: 'none' }} onChange={handleUploadYAML} />
        </Button>
      </LoadWrapper>
    </Box>
  )
}

export default function NewExperiment() {
  const history = useHistory()
  const dispatch = useStoreDispatch()

  const [initialValues, setInitialValues] = useState<Experiment>(defaultExperimentSchema)

  const handleOnSubmit = (values: Experiment, actions: FormikHelpers<Experiment>) => {
    const parsedValues = parseSubmit(values)

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
  }

  return (
    <StepperProvider>
      <Formik
        enableReinitialize
        initialValues={initialValues}
        validationSchema={validationSchema}
        onSubmit={handleOnSubmit}
      >
        <Paper variant="outlined" style={{ height: '100%' }}>
          <PaperTop title="New Experiment" />
          <Grid container>
            <Grid item xs={12} sm={8}>
              <Form>
                <Stepper />
              </Form>
            </Grid>
            <Grid item xs={12} sm={4}>
              <Actions setInitialValues={setInitialValues} />
            </Grid>
          </Grid>
        </Paper>
      </Formik>
    </StepperProvider>
  )
}
