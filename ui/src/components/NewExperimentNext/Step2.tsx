import { Box, Button, Divider, Grid, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'
import basicData, { schema } from './data/basic'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import { setBasic, setStep2 } from 'slices/experiments'
import { useEffect, useMemo, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import AdvancedOptions from 'components/AdvancedOptions'
import CheckIcon from '@material-ui/icons/Check'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import Scheduler from './form/Scheduler'
import Scope from './form/Scope'
import SkeletonN from 'components-mui/SkeletonN'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import { string as yupString } from 'yup'

const useStyles = makeStyles((theme) =>
  createStyles({
    submit: {
      borderColor: theme.palette.success.main,
    },
    submitIcon: {
      color: theme.palette.success.main,
    },
    asButton: {
      cursor: 'pointer',
    },
  })
)

interface Step2Props {
  inWorkflow?: boolean
}

const Step2: React.FC<Step2Props> = ({ inWorkflow = false }) => {
  const classes = useStyles()

  const { namespaces, step2, basic, target } = useStoreSelector((state) => state.experiments)
  const scopeDisabled = target.kind === 'AwsChaos' || target.kind === 'GcpChaos'
  const dispatch = useStoreDispatch()

  const originalInit = useMemo(() => (inWorkflow ? { ...basicData, duration: '' } : basicData), [inWorkflow])
  const [init, setInit] = useState(originalInit)

  useEffect(() => {
    setInit({
      ...originalInit,
      ...basic,
    })
  }, [originalInit, basic])

  const handleOnSubmitStep2 = (values: Record<string, any>) => {
    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep2', values)
    }

    dispatch(setBasic(values))
    dispatch(setStep2(true))
  }

  const handleUndo = () => dispatch(setStep2(false))

  return (
    <Paper className={step2 ? classes.submit : ''}>
      <Box display="flex" justifyContent="space-between" mb={step2 ? 0 : 6}>
        <Box display="flex" alignItems="center">
          {step2 && (
            <Box display="flex" mr={3}>
              <CheckIcon className={classes.submitIcon} />
            </Box>
          )}
          <Typography>{T('newE.titleStep2')}</Typography>
        </Box>
        {step2 && <UndoIcon className={classes.asButton} onClick={handleUndo} />}
      </Box>
      <Box position="relative" hidden={step2}>
        <Formik
          enableReinitialize
          initialValues={init}
          validationSchema={
            inWorkflow
              ? schema.shape({
                  duration: yupString().required('The duration is required'),
                })
              : schema
          }
          validateOnChange={false}
          onSubmit={handleOnSubmitStep2}
        >
          {({ errors, touched }) => (
            <Form>
              <Grid container spacing={6}>
                <Grid item xs={6}>
                  <Box mb={3}>
                    <Typography color={scopeDisabled ? 'textSecondary' : undefined}>
                      {T('newE.steps.scope')}
                      {scopeDisabled && T('newE.steps.scopeDisabled')}
                    </Typography>
                  </Box>
                  {namespaces.length ? <Scope namespaces={namespaces} /> : <SkeletonN n={6} />}
                </Grid>
                <Grid item xs={6}>
                  <Box mb={3}>
                    <Typography>{T('newE.steps.basic')}</Typography>
                  </Box>
                  <TextField
                    fast
                    name="name"
                    label={T('common.name')}
                    helperText={errors.name && touched.name ? errors.name : T('newE.basic.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />
                  {inWorkflow && (
                    <TextField
                      fast
                      name="duration"
                      label={T('newE.schedule.duration')}
                      helperText={
                        (errors as any).duration && (touched as any).duration
                          ? (errors as any).duration
                          : T('newW.node.durationHelper')
                      }
                      error={(errors as any).duration && (touched as any).duration ? true : false}
                    />
                  )}

                  <AdvancedOptions>
                    {namespaces.length && (
                      <SelectField
                        name="namespace"
                        label={T('k8s.namespace')}
                        helperText={T('newE.basic.namespaceHelper')}
                      >
                        {namespaces.map((n) => (
                          <MenuItem key={n} value={n}>
                            {n}
                          </MenuItem>
                        ))}
                      </SelectField>
                    )}
                    <LabelField name="labels" label={T('k8s.labels')} isKV />
                    <LabelField name="annotations" label={T('k8s.annotations')} isKV />
                  </AdvancedOptions>
                  <Box my={3}>
                    <Divider />
                  </Box>
                  <Scheduler errors={errors} touched={touched} />
                  <Box mt={6} textAlign="right">
                    <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
                      {T('common.submit')}
                    </Button>
                  </Box>
                </Grid>
              </Grid>
            </Form>
          )}
        </Formik>
      </Box>
    </Paper>
  )
}

export default Step2
