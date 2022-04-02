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
import { Box, Button, MenuItem, StepLabel, Typography } from '@mui/material'
import NewExperimentNext, { NewExperimentHandles } from 'components/NewExperimentNext'
import { SelectField, TextField } from 'components/FormField'
import { Template, TemplateType, setTemplate, updateTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import AddCircleIcon from '@mui/icons-material/AddCircle'
import CloseIcon from '@mui/icons-material/Close'
import { Formik } from 'formik'
import HTTPTask from './HTTPTask'
import SerialOrParallel from './SerailOrParallel'
import Space from '@ui/mui-extends/esm/Space'
import Suspend from './Suspend'
import Task from './Task'
import i18n from 'components/T'
import { makeStyles } from '@mui/styles'
import { parseYAML } from 'lib/formikhelpers'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'

const useStyles = makeStyles({
  field: {
    width: 180,
  },
})

export enum RenderableTemplateType {
  HTTP = 'http',
}

export type AllTemplateType = RenderableTemplateType | TemplateType

const types = Object.values({ ...TemplateType, ...RenderableTemplateType })

interface AddProps {
  childIndex?: number
  parentTemplates?: Template[]
  setParentTemplates?: React.Dispatch<React.SetStateAction<Template[]>>
  setParentExpand?: React.Dispatch<React.SetStateAction<number>>
  externalTemplate?: Template
  update?: number
  updateCallback?: () => void
}

const Add: React.FC<AddProps> = ({
  childIndex,
  parentTemplates,
  setParentTemplates,
  setParentExpand,
  externalTemplate,
  update,
  updateCallback,
}) => {
  const classes = useStyles()
  const intl = useIntl()

  const dispatch = useStoreDispatch()
  const { templates: storeTemplates } = useStoreSelector((state) => state.workflows)

  const [initialValues, setInitialValues] = useState({
    type: TemplateType.Single as AllTemplateType,
    num: 2,
    name: '',
    deadline: '',
    container: {
      name: '',
      image: '',
      command: [] as string[],
    },
    conditionalBranches: [
      {
        target: '',
        expression: '',
      },
      {
        target: '',
        expression: '',
      },
    ],
  })
  const [num, setNum] = useState(-1)
  const [templates, setTemplates] = useState<Template[]>([])
  const formRef = useRef<any>()
  const newERef = useRef<NewExperimentHandles>(null)
  const [typeOfTemplate, setTypeOfTemplate] = useState<AllTemplateType>(TemplateType.Single)

  // use methods instead of this state
  const [isRenderedHTTPTask, setIsRenderedHTTPTask] = useState(false)

  const fillExperiment = (t: Template) => {
    const e = t.experiment

    const { kind, basic, spec } = parseYAML(e)

    dispatch(
      setExternalExperiment({
        kindAction: [kind, spec.action ?? ''],
        spec,
        basic,
      })
    )
  }

  useEffect(() => {
    // TODO: use API provided by server
    if (externalTemplate) {
      if (
        externalTemplate.type === 'custom' &&
        externalTemplate.custom &&
        externalTemplate.custom.container.name.endsWith('-rendered-http-request')
      ) {
        setIsRenderedHTTPTask(true)
      }
    } else {
      setIsRenderedHTTPTask(false)
    }
  }, [externalTemplate])

  useEffect(() => {
    if (externalTemplate) {
      const { type, name, deadline, children, custom } = externalTemplate

      switch (type) {
        case 'single':
          fillExperiment(externalTemplate)

          break
        case 'serial':
        case 'parallel':
        case 'custom':
          const templates = children!

          setTemplates(templates)
          setNum(templates.length)

          // TODO: if rendered http set type to http

          break
      }

      setInitialValues({
        ...initialValues,
        type,
        num: children ? children.length : 2,
        name,
        deadline: deadline || '',
        ...custom,
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [externalTemplate])

  const resetNoSingle = () => {
    setNum(-1)
    setTemplates([])
  }

  const onValidate = ({ type, num: newNum }: { type: string; num: number }) => {
    setIsRenderedHTTPTask(type === 'http')
    setTypeOfTemplate(type as AllTemplateType)

    const prevType = formRef.current.values.type

    if (prevType !== 'single' && type === 'single') {
      resetNoSingle()

      return
    }

    if (type === 'serial' || type === 'parallel' || type === 'custom') {
      if (typeof newNum !== 'number' || newNum < 0) {
        if (isRenderedHTTPTask) {
          resetNoSingle()
          return
        }

        formRef.current.setFieldValue('num', 2)

        return
      }

      // Protect exist templates
      if (newNum < templates.length) {
        return
      }

      setNum(newNum)

      return
    }

    if (type === 'http') {
      resetNoSingle()
      return
    }

    if (type === 'suspend') {
      if (prevType === 'serial' || prevType === 'parallel' || prevType === 'custom') {
        resetNoSingle()
      }
    }
  }

  const submit = (template: Template) => {
    if (storeTemplates.some((t) => t.name === template.name)) {
      dispatch(
        setAlert({
          type: 'warning',
          message: i18n('newW.messages.redundant', intl),
        })
      )

      return
    }

    if (childIndex !== undefined) {
      if (parentTemplates![childIndex!]) {
        const tmp = JSON.parse(JSON.stringify(parentTemplates!))
        tmp[childIndex!] = template

        setParentTemplates!(tmp)
      } else {
        setParentTemplates!([...parentTemplates!, template])
      }

      setParentExpand!(-1)
    } else {
      dispatch(update !== undefined ? updateTemplate({ ...template, index: update }) : setTemplate(template))
      typeof updateCallback === 'function' && updateCallback()
    }
  }

  const onSubmit = (experiment: any) => {
    const type = formRef.current.values.type

    const name = experiment.metadata.name
    const template = {
      type,
      name,
      experiment,
    }

    submit(template)

    dispatch(resetNewExperiment())
  }

  const submitNoSingleNode = () => {
    const { type, name, deadline, container, conditionalBranches } = formRef.current.values
    const template: Template = {
      type,
      name: name.trim(),
      deadline,
      children: templates,
    }
    if (type === 'custom') {
      template.custom = {
        container,
        conditionalBranches,
      }
    }

    submit(template)

    resetNoSingle()
  }

  return (
    <>
      <Formik
        innerRef={formRef}
        initialValues={initialValues}
        enableReinitialize
        onSubmit={submitNoSingleNode}
        validate={onValidate}
        validateOnBlur={false}
      >
        {({ values, setFieldValue, errors, touched }) => {
          return (
            <>
              <StepLabel icon={<AddCircleIcon color="primary" />}>
                <Space direction="row">
                  <SelectField className={classes.field} name="type" label={i18n('newW.node.choose')}>
                    {types.map((d) => (
                      <MenuItem key={d} value={d}>
                        <Typography variant="body2">{i18n(`newW.node.${d}`)}</Typography>
                      </MenuItem>
                    ))}
                  </SelectField>
                  {num > 0 && (
                    <TextField
                      className={classes.field}
                      type="number"
                      name="num"
                      label={i18n('newW.node.number')}
                      inputProps={{ min: 1 }}
                    />
                  )}
                  {update !== undefined && (
                    <Button variant="outlined" startIcon={<CloseIcon />} onClick={updateCallback}>
                      {i18n('common.cancelEdit')}
                    </Button>
                  )}
                </Space>
              </StepLabel>

              {(values.type === 'serial' || values.type === 'parallel') && (
                <SerialOrParallel
                  name={values.name}
                  deadline={values.deadline}
                  type={values.type as TemplateType}
                  childrenCount={values.num}
                  submitTemplate={submit}
                  templates={templates}
                ></SerialOrParallel>
              )}
              {values.type === 'custom' && !isRenderedHTTPTask && (
                <>
                  <Task
                    name={values.name}
                    deadline={values.deadline}
                    type={values.type as TemplateType}
                    childrenCount={values.num}
                    container={values.container}
                    submitTemplate={submit}
                    templates={templates}
                    conditionalBranches={values.conditionalBranches}
                  ></Task>
                </>
              )}
              {isRenderedHTTPTask && (
                <Box mt={3}>
                  <HTTPTask
                    name={values.name}
                    type={values.type as TemplateType}
                    childrenCount={values.num}
                    templates={templates}
                    submitTemplate={submit}
                    externalTemplate={externalTemplate}
                  />
                </Box>
              )}
            </>
          )
        }}
      </Formik>

      {num < 0 && (
        <Box ml={8}>
          {typeOfTemplate === 'suspend' && (
            <Box mt={3}>
              <Suspend initialValues={initialValues} submit={submit} />
            </Box>
          )}

          {typeOfTemplate === 'single' && (
            <Box display="initial">
              <NewExperimentNext ref={newERef} onSubmit={onSubmit} inWorkflow={true} />
            </Box>
          )}
        </Box>
      )}
    </>
  )
}

export default Add
