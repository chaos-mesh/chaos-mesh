import { Box, Button, Divider, Grid, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'
import {
  Fields as ScheduleSpecificFields,
  data as scheduleSpecificData,
  schema as scheduleSpecificSchema,
} from 'components/Schedule/types'
import basicData, { schema } from './data/basic'
import { setBasic, setScheduleSpecific, setStep2 } from 'slices/experiments'
import { useEffect, useMemo, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import AdvancedOptions from 'components/AdvancedOptions'
import CheckIcon from '@material-ui/icons/Check'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import Scheduler from './form/Scheduler'
import Scope from './form/Scope'
import SkeletonN from 'components-mui/SkeletonN'
import Space from 'components-mui/Space'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import { string as yupString } from 'yup'

function isInstant(target: any) {
  if (
    target.kind === 'PodChaos' &&
    (target.pod_chaos.action === 'pod-kill' || target.pod_chaos.action === 'container-kill')
  ) {
    return true
  }

  return false
}

interface Step2Props {
  inWorkflow?: boolean
  inSchedule?: boolean
}

const Step2: React.FC<Step2Props> = ({ inWorkflow = false, inSchedule = false }) => {
  const { namespaces, step2, basic, target, scheduleSpecific } = useStoreSelector((state) => state.experiments)
  const scopeDisabled = target.kind === 'AwsChaos' || target.kind === 'GcpChaos'
  const dispatch = useStoreDispatch()

  const originalInit = useMemo(
    () =>
      inWorkflow
        ? { ...basicData, scheduler: undefined, deadline: '' }
        : inSchedule
        ? { ...basicData, ...scheduleSpecificData }
        : basicData,
    [inWorkflow, inSchedule]
  )
  const [init, setInit] = useState(originalInit)

  useEffect(() => {
    setInit({
      ...originalInit,
      ...basic,
      ...scheduleSpecific,
    })
  }, [originalInit, basic, scheduleSpecific])

  const handleOnSubmitStep2 = (_values: Record<string, any>) => {
    const values = schema.cast(_values) as Record<string, any>

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep2', values)
    }

    if (inSchedule) {
      dispatch(
        setScheduleSpecific({
          schedule: values.schedule,
          starting_deadline_seconds: values.starting_deadline_seconds,
          concurrency_policy: values.concurrency_policy,
          history_limit: values.history_limit,
        })
      )

      delete values.schedule
      delete values.starting_deadline_seconds
      delete values.concurrency_policy
      delete values.history_limit
    }

    dispatch(setBasic(values))
    dispatch(setStep2(true))
  }

  const handleUndo = () => dispatch(setStep2(false))

  return (
    <Paper sx={{ borderColor: step2 ? 'success.main' : undefined }}>
      <Box display="flex" justifyContent="space-between" mb={step2 ? 0 : 6}>
        <Box display="flex" alignItems="center">
          {step2 && (
            <Box display="flex" mr={3}>
              <CheckIcon sx={{ color: 'success.main' }} />
            </Box>
          )}
          <Typography>{T(`${inSchedule ? 'newS' : 'newE'}.titleStep2`)}</Typography>
        </Box>
        {step2 && <UndoIcon onClick={handleUndo} sx={{ cursor: 'pointer' }} />}
      </Box>
      <Box position="relative" hidden={step2}>
        <Formik
          enableReinitialize
          initialValues={init}
          validationSchema={
            inWorkflow
              ? schema.shape({
                  deadline: yupString().required('The deadline is required'),
                })
              : inSchedule
              ? schema.shape(scheduleSpecificSchema)
              : schema
          }
          validateOnChange={false}
          onSubmit={handleOnSubmitStep2}
        >
          {({ errors, touched }) => (
            <Form>
              <Grid container spacing={6}>
                <Grid item xs={6}>
                  <Space>
                    <Typography color={scopeDisabled ? 'textSecondary' : undefined}>
                      {T('newE.steps.scope')}
                      {scopeDisabled && T('newE.steps.scopeDisabled')}
                    </Typography>
                    {namespaces.length ? <Scope namespaces={namespaces} /> : <SkeletonN n={6} />}
                  </Space>
                </Grid>
                <Grid item xs={6}>
                  <Space>
                    <Typography>{T('newE.steps.basic')}</Typography>
                    <TextField
                      fast
                      name="name"
                      label={T('common.name')}
                      helperText={
                        errors.name && touched.name
                          ? errors.name
                          : T(`${inSchedule ? 'newS' : 'newE'}.basic.nameHelper`)
                      }
                      error={errors.name && touched.name ? true : false}
                    />
                    {inWorkflow && (
                      <TextField
                        fast
                        name="deadline"
                        label={T('newW.node.deadline')}
                        helperText={
                          (errors as any).deadline && (touched as any).deadline
                            ? (errors as any).deadline
                            : T('newW.node.deadlineHelper')
                        }
                        error={(errors as any).deadline && (touched as any).deadline ? true : false}
                      />
                    )}
                    {inSchedule && <ScheduleSpecificFields errors={errors} touched={touched} />}
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
                    {!inWorkflow && !isInstant(target) && (
                      <>
                        <Divider />
                        <Scheduler errors={errors} touched={touched} inSchedule={inSchedule} />
                      </>
                    )}
                  </Space>
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
