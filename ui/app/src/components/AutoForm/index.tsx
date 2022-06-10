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
import { Box, Divider, FormHelperText, MenuItem, Typography } from '@mui/material'
import { eval as expEval, parse } from 'expression-eval'
import { Form, Formik, getIn } from 'formik'
import type { FormikConfig, FormikErrors, FormikTouched, FormikValues } from 'formik'
import { useEffect, useState } from 'react'

import Checkbox from '@ui/mui-extends/esm/Checkbox'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import { AutocompleteField, SelectField, Submit, TextField, TextTextField } from 'components/FormField'
import Scope from 'components/Scope'
import { T } from 'components/T'

import { concatKindAction } from 'lib/utils'

import Info from './Info'
import Schedule from './Schedule'
import { removeScheduleValues, scheduleInitialValues, scopeInitialValues, workflowNodeInfoInitialValues } from './data'
import { chooseSchemaByBelong } from './validation'

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
  const [initialValues, setInitialValues] = useState<Record<string, any>>({
    id,
    kind,
    action,
    ...(kind === 'NetworkChaos' && { target: scopeInitialValues }),
    ...(kind !== 'PhysicalMachineChaos' && kind !== 'Suspend' && scopeInitialValues),
    ...(belong === Belong.Workflow && { ...workflowNodeInfoInitialValues, templateType: kind }),
  })
  const [form, setForm] = useState<AtomFormData[]>([])
  const [scheduled, setScheduled] = useState(false)

  const { namespaces } = useStoreSelector((state) => state.experiments)

  useEffect(() => {
    function formToRecords(form: AtomFormData[]) {
      return form.reduce((acc, d) => {
        if (d.field === 'ref') {
          acc[d.label] = formToRecords(d.children!)
        } else {
          acc[d.label] = d.value
        }

        return acc
      }, {} as Record<string, any>)
    }

    async function loadData() {
      if (kind === 'Suspend') {
        setInitialValues((oldValues) => ({
          ...oldValues,
          ...formikProps.initialValues,
        }))

        return
      }

      const { data }: { data: AtomFormData[] } = await import(`../../formik/${kind}.ts`)
      const form = action
        ? data.filter((d) => {
            // Since some selectors are not yet supported, ignore target for now.
            if (kind === 'NetworkChaos' && d.label === 'target') {
              return false
            }

            if (d.when) {
              const parsed = parse(d.when)

              return expEval(parsed, { action })
            }

            return true
          })
        : data

      setInitialValues((oldValues) => ({
        ...oldValues,
        ...(formikProps.initialValues || formToRecords(form)),
      }))
      setForm(form)
    }

    if (kind) {
      loadData()
    }
  }, [kind, action, formikProps.initialValues])

  const renderForm = (
    form: AtomFormData[],
    errors: FormikErrors<Record<string, any>>,
    touched: FormikTouched<Record<string, any>>,
    parent?: string
  ): any[] => {
    // eslint-disable-next-line array-callback-return
    return form.map(({ field, label, items, helperText, children, multiple }) => {
      const touch = getIn(touched, label)
      const error = getIn(errors, label)
      const touchAndError = touch && error

      const _label = parent ? [parent, label].join('.') : label

      switch (field) {
        case 'text':
          return (
            <TextField
              key={_label}
              name={_label}
              label={label}
              helperText={touchAndError ? error : helperText}
              error={touchAndError}
            />
          )
        case 'number':
          return (
            <TextField
              type="number"
              key={_label}
              name={_label}
              label={label}
              helperText={touchAndError ? error : helperText}
              error={touchAndError}
            />
          )
        case 'select':
          return (
            <SelectField
              key={_label}
              name={_label}
              label={label}
              helperText={touchAndError ? error : helperText}
              error={touchAndError}
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
              helperText={touchAndError ? error : helperText}
              error={touchAndError}
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
          return (
            <Box key={_label}>
              <Typography variant="body2" fontWeight={500} mb={3}>
                {label}
              </Typography>
              {helperText && <FormHelperText>{helperText}</FormHelperText>}
              <Space direction="row">
                <Divider orientation="vertical" flexItem />
                <Space>{renderForm(children!, errors, touched, _label)}</Space>
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
      {({ errors, touched }) => (
        <Form>
          <Space>
            <Typography variant="h6" fontWeight="bold">
              {concatKindAction(kind, action)}
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
            {renderForm(form, errors, touched)}
            {kind === 'NetworkChaos' && (
              <>
                <Typography variant="body2" fontWeight={500} mb={3}>
                  target
                </Typography>
                <Space direction="row">
                  <Divider orientation="vertical" flexItem />
                  <Scope
                    namespaces={namespaces}
                    scope="target.selector"
                    modeScope="target"
                    podsPreviewTitle={<T id="newE.target.network.target.podsPreview" />}
                    podsPreviewDesc={<T id="newE.target.network.target.podsPreviewHelper" />}
                  />
                </Space>
              </>
            )}
            {kind !== 'Suspend' && (
              <>
                <Divider />
                <Typography variant="h6" fontWeight="bold">
                  <T id="newE.steps.scope" />
                </Typography>
                {kind !== 'PhysicalMachineChaos' && <Scope kind={kind} namespaces={namespaces} />}
                <Divider />
                <Box>
                  <Typography variant="h6" fontWeight="bold">
                    Schedule
                  </Typography>
                  <Checkbox
                    label="Scheduled"
                    helperText="Check the box to convert the Experiment into a Schedule."
                    checked={scheduled}
                    onChange={switchToSchedule}
                  />
                </Box>
                {scheduled && <Schedule />}
                <Divider />
              </>
            )}
            <Typography variant="h6" fontWeight="bold">
              Info
            </Typography>
            <Info belong={belong} kind={kind} action={action} />
            <Submit />
          </Space>
        </Form>
      )}
    </Formik>
  )
}

export default AutoForm
