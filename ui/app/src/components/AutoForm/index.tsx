/*
 * Copyright 2022 Chaos Mesh Authors.
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
import AddIcon from '@mui/icons-material/Add'
import { Box, Button, Chip, Divider, FormHelperText, MenuItem, Typography } from '@mui/material'
import { eval as expEval, parse } from 'expression-eval'
import { Form, Formik, FormikProps, getIn } from 'formik'
import type { FormikConfig, FormikValues } from 'formik'
import _ from 'lodash'
import { useGetCommonChaosAvailableNamespaces } from 'openapi'
import { Fragment, useEffect, useState } from 'react'

import Checkbox from '@ui/mui-extends/esm/Checkbox'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import { AutocompleteField, SelectField, Submit, TextField, TextTextField } from 'components/FormField'
import { SpecialTemplateType } from 'components/NewWorkflowNext/utils/convert'
import Scope from 'components/Scope'
import Mode from 'components/Scope/Mode'
import { T } from 'components/T'

import { concatKindAction } from 'lib/utils'

import Info from './Info'
import Schedule from './Schedule'
import { removeScheduleValues, scheduleInitialValues, scopeInitialValues, workflowNodeInfoInitialValues } from './data'
import { chooseSchemaByBelong } from './validation'
import { Stale } from 'api/queryUtils'

export enum Belong {
  Experiment = 'Experiment',
  Schedule = 'Schedule',
  Workflow = 'Workflow',
}

export interface AutoFormProps {
  belong?: Belong
  id?: uuid
  kind: string // e.g. 'PodChaos'
  act?: string // e.g. 'pod-failure'
  formikProps: Partial<Pick<FormikConfig<FormikValues>, 'initialValues' | 'onSubmit'>>
}

export interface AtomFormData {
  field: 'text' | 'number' | 'select' | 'label' | 'numbers' | 'text-text' | 'text-label' | 'ref'
  label: string
  value: string
  items?: any[]
  children?: AtomFormData[]
  multiple?: boolean
  helperText?: string
  when?: string
}

const AutoForm: React.FC<AutoFormProps> = ({ belong = Belong.Experiment, id, kind, act: action, formikProps }) => {
  const kindAction = concatKindAction(kind, action)

  const { useNewPhysicalMachine } = useStoreSelector((state) => state.settings)
  const noScope =
    kind === SpecialTemplateType.Suspend || kind === SpecialTemplateType.Serial || kind === SpecialTemplateType.Parallel

  const [initialValues, setInitialValues] = useState<FormikValues>({
    id,
    kind,
    action,
    ...(kind === 'NetworkChaos' && { target: scopeInitialValues({ hasSelector: true }) }),
    ...(!noScope &&
      scopeInitialValues({ hasSelector: kind === 'PhysicalMachineChaos' && !useNewPhysicalMachine ? false : true })),
    ...(belong === Belong.Workflow && { ...workflowNodeInfoInitialValues, templateType: kind }),
  })
  const hasSelector = !!initialValues.selector
  const [form, setForm] = useState<AtomFormData[]>([])
  const [scheduled, setScheduled] = useState(false)

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  useEffect(() => {
    function formToRecords(form: AtomFormData[]) {
      return form.reduce((acc, d) => {
        if (d.field === 'ref') {
          acc[d.label] = d.multiple ? [formToRecords(d.children!)] : formToRecords(d.children!)
        } else {
          acc[d.label] = d.value
        }

        return acc
      }, {} as FormikValues)
    }

    async function loadData() {
      if (
        kind === SpecialTemplateType.Suspend ||
        kind === SpecialTemplateType.Serial ||
        kind === SpecialTemplateType.Parallel
      ) {
        setInitialValues((oldValues) => _.merge({}, oldValues, formikProps.initialValues))

        return
      }

      const { data }: { data: AtomFormData[] } = await import(`../../formik/${kind}.ts`)
      const form = action
        ? data.filter((d) => {
            // Since some selectors are not yet supported, ignore target for now.
            if (kind === 'NetworkChaos' && d.label === 'target') {
              return false
            }

            if (kind === 'PhysicalMachineChaos' && useNewPhysicalMachine && d.label === 'address') {
              return false
            }

            if (d.when) {
              const parsed = parse(d.when)

              return expEval(parsed, { action })
            }

            return true
          })
        : data

      setInitialValues((oldValues) => _.merge({}, oldValues, formikProps.initialValues || formToRecords(form)))
      setForm(form)
    }

    if (kind) {
      loadData()
    }
  }, [kind, action, useNewPhysicalMachine, formikProps.initialValues])

  const renderForm = (
    form: AtomFormData[],
    props: FormikProps<FormikValues>,
    parent?: string,
    index?: number
  ): any[] => {
    const { values, errors, touched, setFieldValue } = props

    // eslint-disable-next-line array-callback-return
    return form.map(({ field, label, items, helperText, children, multiple }) => {
      const error = getIn(errors, label)
      const touch = getIn(touched, label)
      const errorAndTouch = error && touch

      let _label = label
      if (parent) {
        if (index !== undefined) {
          _label = `${parent}[${index}].${label}`
        } else {
          _label = `${parent}.${label}`
        }
      }

      switch (field) {
        case 'text':
          return (
            <TextField
              key={_label}
              name={_label}
              label={label}
              helperText={errorAndTouch ? error : helperText}
              error={errorAndTouch}
            />
          )
        case 'number':
          return (
            <TextField
              type="number"
              key={_label}
              name={_label}
              label={label}
              helperText={errorAndTouch ? error : helperText}
              error={errorAndTouch}
            />
          )
        case 'select':
          return (
            <SelectField
              key={_label}
              name={_label}
              label={label}
              helperText={errorAndTouch ? error : helperText}
              error={errorAndTouch}
            >
              {items!.map((option) => (
                <MenuItem key={option.toString()} value={option}>
                  {option.toString()}
                </MenuItem>
              ))}
            </SelectField>
          )
        case 'label':
          return (
            <AutocompleteField
              freeSolo
              multiple
              key={_label}
              name={_label}
              label={label}
              helperText={errorAndTouch ? error : helperText}
              error={errorAndTouch}
              options={[]}
            />
          )
        case 'text-text':
        case 'text-label':
          return (
            <TextTextField
              key={_label}
              name={_label}
              label={label}
              helperText={helperText}
              valueLabeled={field === 'text-label'}
            />
          )
        case 'ref':
          const value = getIn(values, _label)
          const isMultiple = multiple && _.isArray(value)

          return (
            <Box key={_label}>
              <Typography fontWeight={500} mb={3}>
                {label}
              </Typography>
              {helperText && <FormHelperText>{helperText}</FormHelperText>}
              <Space direction="row">
                <Divider orientation="vertical" flexItem />
                <Space>
                  {isMultiple
                    ? value.map((_, index) => (
                        <Fragment key={index}>
                          <Box>
                            <Chip
                              label={`Item ${index + 1}`}
                              color="primary"
                              size="small"
                              onDelete={() => {
                                setFieldValue(
                                  _label,
                                  value.filter((_, i) => i !== index)
                                )
                              }}
                            />
                          </Box>
                          {renderForm(children!, props, _label, index)}
                        </Fragment>
                      ))
                    : renderForm(children!, props, _label)}
                  {isMultiple && (
                    <Box>
                      <Button
                        variant="contained"
                        startIcon={<AddIcon />}
                        onClick={() => {
                          setFieldValue(_label, [...value, _.mapValues(value[0], () => '')])
                        }}
                      >{`Add ${label} Item`}</Button>
                    </Box>
                  )}
                </Space>
              </Space>
            </Box>
          )
      }
    })
  }

  const switchToSchedule = () => {
    if (!scheduled) {
      setInitialValues((oldValues) => ({
        ...oldValues,
        scheduled: true,
        ...scheduleInitialValues,
      }))
      setTimeout(() => setScheduled(true))
    } else {
      setScheduled(false)
      setInitialValues((oldValues) => removeScheduleValues(oldValues))
    }
  }

  return (
    <Formik
      enableReinitialize
      initialValues={initialValues}
      validationSchema={chooseSchemaByBelong(belong, kind, action)}
      onSubmit={formikProps.onSubmit!}
    >
      {(props) => (
        <Form>
          <Space>
            <Typography variant="h6" fontWeight="bold">
              {kindAction}
            </Typography>
            {action && (
              <SelectField
                name="action"
                label="action"
                helperText="The action of the experiment. Automatic filling."
                disabled
              >
                <MenuItem value={action}>{action}</MenuItem>
              </SelectField>
            )}
            {renderForm(form, props)}
            {kind === 'NetworkChaos' && (
              <>
                <Typography fontWeight={500} mb={3}>
                  target
                </Typography>
                <Space direction="row">
                  <Divider orientation="vertical" flexItem />
                  <Scope
                    env="k8s"
                    kind={kind}
                    namespaces={namespaces!}
                    scope="target.selector"
                    modeScope="target"
                    previewTitle={<T id="newE.target.network.target.podsPreview" />}
                  />
                </Space>
              </>
            )}
            {kind !== SpecialTemplateType.Suspend &&
              kind !== SpecialTemplateType.Serial &&
              kind !== SpecialTemplateType.Parallel && (
                <>
                  <Divider />
                  <Typography variant="h6">
                    <T id="newE.steps.scope" />
                  </Typography>
                  {hasSelector ? (
                    <Scope
                      env={kind === 'PhysicalMachineChaos' ? 'physic' : 'k8s'}
                      kind={kind}
                      namespaces={namespaces!}
                    />
                  ) : (
                    <Mode scope="selector" modeScope="" />
                  )}
                  <Divider />
                  <Box>
                    <Typography variant="h6">Schedule</Typography>
                    <Checkbox
                      label="Scheduled"
                      helperText={`Check the box to convert ${kindAction} into a Schedule.`}
                      checked={scheduled}
                      onChange={switchToSchedule}
                    />
                  </Box>
                  {scheduled && <Schedule />}
                  <Divider />
                </>
              )}
            <Typography variant="h6">Info</Typography>
            <Info belong={belong} kind={kind} action={action} />
            <Submit />
          </Space>
        </Form>
      )}
    </Formik>
  )
}

export default AutoForm
