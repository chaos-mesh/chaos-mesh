import {
  Box,
  Button,
  Divider,
  FormControlLabel,
  Grid,
  Paper,
  Radio,
  RadioGroup,
  Snackbar,
  Typography,
} from '@material-ui/core'
import { Form, Formik } from 'formik'
import { IntlShape, useIntl } from 'react-intl'
import React, { useEffect, useState } from 'react'
import { defaultExperimentSchema, validationSchema } from './constants'
import { parseSubmit, yamlToExperiment } from 'lib/formikhelpers'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import Alert from '@material-ui/lab/Alert'
import { Archive } from 'api/archives.type'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import { Experiment } from './types'
import { Experiment as ExperimentResponse } from 'api/experiments.type'
import PaperTop from 'components/PaperTop'
import SkeletonN from 'components/SkeletonN'
import Stepper from './Stepper'
import { StepperProvider } from './Context'
import T from 'components/T'
import api from 'api'
import flat from 'flat'
import { setNeedToRefreshExperiments } from 'slices/experiments'
import { useHistory } from 'react-router-dom'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const Skeleton3 = () => <SkeletonN n={3} />

const LoadWrapper: React.FC<{ title: string | JSX.Element }> = ({ title, children }) => (
  <Box mb={6}>
    <Box mb={6}>
      <Typography>{title}</Typography>
    </Box>
    <Box maxHeight="300px" style={{ overflowY: 'scroll' }}>
      {children}
    </Box>
  </Box>
)

const CustomRadioLabel = (e: ExperimentResponse | Archive) => (
  <Box display="flex" justifyContent="space-between" alignItems="center">
    <Typography variant="body1" component="div">
      {e.name}
    </Typography>
    <Box ml={3}>
      <Typography variant="body2" color="textSecondary">
        {e.uid}
      </Typography>
    </Box>
  </Box>
)

interface ActionsProps {
  setInitialValues: (initialValues: Experiment) => void
  intl: IntlShape
}

const Actions = ({ setInitialValues, intl }: ActionsProps) => {
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
      .then(({ data }) => setInitialValues(yamlToExperiment(data.yaml)))
      .catch(console.log)
  }

  const onArchiveRadioChange = (e: any) => {
    const uuid = e.target.value

    setArchiveRadio(uuid)
    setExperimentRadio('')

    api.archives
      .detail(uuid)
      .then(({ data }) => setInitialValues(yamlToExperiment(data.yaml)))
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
        if (process.env.NODE_ENV === 'development') {
          console.debug('Debug yamlToExperiment:', y)
        }
        setInitialValues(y)
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.importSuccessfully' }),
          })
        )
      } catch (e) {
        console.error(e)
        dispatch(
          setAlert({
            type: 'error',
            message: e.message,
          })
        )
      } finally {
        dispatch(setAlertOpen(true))
      }
    }
    reader.readAsText(f)
  }

  return (
    <Box p={6} pt={12}>
      <LoadWrapper title={T('newE.loadFromExistingExperiments')}>
        <RadioGroup value={experimentRadio} onChange={onExperimentRadioChange}>
          {experiments && experiments.length > 0 ? (
            experiments.map((e) => (
              <FormControlLabel
                key={e.uid}
                value={e.uid}
                control={<Radio color="primary" />}
                label={CustomRadioLabel(e)}
              />
            ))
          ) : experiments?.length === 0 ? (
            <Typography variant="body2">{T('experiments.noExperimentsFound')}</Typography>
          ) : (
            <Skeleton3 />
          )}
        </RadioGroup>
      </LoadWrapper>

      <Box my={6}>
        <Divider />
      </Box>

      <LoadWrapper title={T('newE.loadFromArchives')}>
        <RadioGroup value={archiveRadio} onChange={onArchiveRadioChange}>
          {archives && archives.length > 0 ? (
            archives.map((a) => (
              <FormControlLabel
                key={a.uid}
                value={a.uid}
                control={<Radio color="primary" />}
                label={CustomRadioLabel(a)}
              />
            ))
          ) : archives?.length === 0 ? (
            <Typography variant="body2">{T('archives.no_archives_found')}</Typography>
          ) : (
            <Skeleton3 />
          )}
        </RadioGroup>
      </LoadWrapper>

      <Box my={6}>
        <Divider />
      </Box>

      <LoadWrapper title={T('newE.loadFromYamlFile')}>
        <Button component="label" variant="outlined" size="small" startIcon={<CloudUploadOutlinedIcon />}>
          {T('common.upload')}
          <input type="file" style={{ display: 'none' }} onChange={handleUploadYAML} />
        </Button>
      </LoadWrapper>
    </Box>
  )
}

export default function NewExperiment() {
  const history = useHistory()

  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const [initialValues, setInitialValues] = useState<Experiment>(defaultExperimentSchema)

  const handleOnSubmit = (values: Experiment) => {
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
            message: intl.formatMessage({ id: 'common.createSuccessfully' }),
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
        validateOnChange={false}
        onSubmit={handleOnSubmit}
      >
        {({ errors }) => {
          const flatErrors: Record<string, string> = flat(errors)

          return (
            <Paper variant="outlined" style={{ height: '100%' }}>
              <PaperTop title={T('newE.create')} />
              <Grid container>
                <Grid item xs={12} sm={8}>
                  <Form>
                    <Stepper />
                  </Form>
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Actions setInitialValues={setInitialValues} intl={intl} />
                </Grid>
              </Grid>

              {Object.keys(flatErrors).length > 0 && (
                <Snackbar
                  anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'center',
                  }}
                  open={true}
                >
                  <Alert severity="error">{Object.values(flatErrors).join('/')}</Alert>
                </Snackbar>
              )}
            </Paper>
          )
        }}
      </Formik>
    </StepperProvider>
  )
}
