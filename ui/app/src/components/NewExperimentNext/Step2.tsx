/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { Box, Button, Divider, Grid, MenuItem, Typography } from '@mui/material'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'
import { Fields as ScheduleSpecificFields, data as scheduleSpecificData } from 'components/Schedule/types'
import basicData, { schema as basicSchema } from './data/basic'
import { setBasic, setStep2 } from 'slices/experiments'
import { useEffect, useMemo, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import CheckIcon from '@mui/icons-material/Check'
import Mode from './form/Mode'
import Nodes from './form/Nodes'
import OtherOptions from 'components/OtherOptions'
import Paper from '@ui/mui-extends/esm/Paper'
import PublishIcon from '@mui/icons-material/Publish'
import Scheduler from './form/Scheduler'
import Scope from './form/Scope'
import SkeletonN from '@ui/mui-extends/esm/SkeletonN'
import Space from '@ui/mui-extends/esm/Space'
import UndoIcon from '@mui/icons-material/Undo'
import _isEmpty from 'lodash.isempty'
import i18n from 'components/T'

interface Step2Props {
  inWorkflow?: boolean
  inSchedule?: boolean
}

const Step2: React.FC<Step2Props> = ({ inWorkflow = false, inSchedule = false }) => {
  const { namespaces, step2, env, kindAction, basic } = useStoreSelector((state) => state.experiments)
  const [kind] = kindAction
  const scopeDisabled = kind === 'AWSChaos' || kind === 'GCPChaos'
  const schema = basicSchema({ env, scopeDisabled, scheduled: inSchedule, needDeadline: inWorkflow })
  const dispatch = useStoreDispatch()
  const originalInit = useMemo(
    () =>
      inSchedule
        ? {
            metadata: basicData.metadata,
            spec: {
              ...basicData.spec,
              ...scheduleSpecificData,
            },
          }
        : basicData,
    [inSchedule]
  )
  const [init, setInit] = useState(originalInit)

  useEffect(() => {
    if (!_isEmpty(basic)) {
      setInit({
        metadata: {
          ...originalInit.metadata,
          ...basic.metadata,
        },
        spec: {
          ...originalInit.spec,
          ...basic.spec,
          selector: {
            ...originalInit.spec.selector,
            ...basic.spec.selector,
          },
        },
      })
    }
  }, [originalInit, basic])

  const handleOnSubmitStep2 = (_values: Record<string, any>) => {
    const values = schema.cast(_values) as Record<string, any>

    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep2:', values)
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
          <Typography>{i18n(`${inSchedule ? 'newS' : 'newE'}.titleStep2`)}</Typography>
        </Box>
        {step2 && <UndoIcon onClick={handleUndo} sx={{ cursor: 'pointer' }} />}
      </Box>
      <Box position="relative" hidden={step2}>
        <Formik
          enableReinitialize
          initialValues={init}
          validationSchema={schema}
          validateOnChange={false}
          onSubmit={handleOnSubmitStep2}
        >
          {({ errors, touched }) => (
            <Form>
              <Grid container spacing={6}>
                <Grid item xs={6}>
                  <Space>
                    <Typography sx={{ color: scopeDisabled ? 'text.disabled' : undefined }}>
                      {i18n('newE.steps.scope')}
                      {scopeDisabled && i18n('newE.steps.scopeDisabled')}
                    </Typography>
                    {env === 'k8s' ? (
                      namespaces.length ? (
                        <Scope namespaces={namespaces} />
                      ) : (
                        <SkeletonN n={6} />
                      )
                    ) : (
                      <>
                        <Nodes />
                        <Divider />
                        <Typography>{i18n('newE.scope.mode')}</Typography>
                        <Mode disabled={false} modeScope={'spec'} scope={'spec.selector'} />
                      </>
                    )}
                  </Space>
                </Grid>
                <Grid item xs={6}>
                  <Space>
                    <Typography>{i18n('newE.steps.basic')}</Typography>
                    <TextField
                      fast
                      name="metadata.name"
                      label={i18n('common.name')}
                      helperText={
                        errors.metadata?.name && touched.metadata?.name
                          ? errors.metadata.name
                          : i18n(`${inSchedule ? 'newS' : 'newE'}.basic.nameHelper`)
                      }
                      error={errors.metadata?.name && touched.metadata?.name ? true : false}
                    />
                    {inWorkflow && (
                      <TextField
                        fast
                        name="spec.duration"
                        label={i18n('newW.node.deadline')}
                        helperText={
                          errors.spec?.duration && touched.spec?.duration
                            ? errors.spec?.duration
                            : i18n('newW.node.deadlineHelper')
                        }
                        error={errors.spec?.duration && touched.spec?.duration ? true : false}
                      />
                    )}
                    {inSchedule && <ScheduleSpecificFields errors={errors} touched={touched} />}
                    <OtherOptions>
                      {namespaces.length && (
                        <SelectField
                          name="metadata.namespace"
                          label={i18n('k8s.namespace')}
                          helperText={i18n('newE.basic.namespaceHelper')}
                        >
                          {namespaces.map((n) => (
                            <MenuItem key={n} value={n}>
                              {n}
                            </MenuItem>
                          ))}
                        </SelectField>
                      )}
                      <LabelField name="metadata.labels" label={i18n('k8s.labels')} isKV />
                      <LabelField name="metadata.annotations" label={i18n('k8s.annotations')} isKV />
                    </OtherOptions>
                    {!inWorkflow && (
                      <>
                        <Divider />
                        <Scheduler errors={errors} touched={touched} inSchedule={inSchedule} />
                      </>
                    )}
                  </Space>
                  <Box mt={6} textAlign="right">
                    <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
                      {i18n('common.submit')}
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
