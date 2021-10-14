import {
  Autocomplete,
  Box,
  Button,
  IconButton,
  TextField as MUITextField,
  MenuItem,
  StepLabel,
  Typography,
} from '@material-ui/core'
import { Branch, Template, TemplateType, setTemplate, updateTemplate } from 'slices/workflows'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, Submit, TextField } from 'components/FormField'
import NewExperimentNext, { NewExperimentHandles } from 'components/NewExperimentNext'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'
import { validateDeadline, validateImage, validateName } from 'lib/formikhelpers'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowRightIcon from '@material-ui/icons/ArrowRight'
import CloseIcon from '@material-ui/icons/Close'
import HTTPTask from './HTTPTask'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import RemoveCircleIcon from '@material-ui/icons/RemoveCircle'
import SerialOrParallel from './SerailOrParallel'
import Space from 'components-mui/Space'
import Suspend from './Suspend'
import T from 'components/T'
import { makeStyles } from '@material-ui/styles'
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
  const [expand, setExpand] = useState(-1)
  const [otherTypes, setOtherTypes] = useState<'suspend' | ''>('')
  const [templates, setTemplates] = useState<Template[]>([])
  const templateNames = [...new Set([...storeTemplates, ...templates].map((t) => t.name))]
  const formRef = useRef<any>()
  const newERef = useRef<NewExperimentHandles>(null)
  const [typeOfTemplate, setTypeOfTemplate] = useState<AllTemplateType>(TemplateType.Single)

  const isRenderedHTTPTask = (): boolean => {
    return typeOfTemplate === RenderableTemplateType.HTTP
  }

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
        case 'suspend':
          setOtherTypes(type)
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
    setTypeOfTemplate(type as AllTemplateType)

    if (type !== 'suspend' && type !== 'http') {
      setOtherTypes('')
    }

    const prevType = formRef.current.values.type

    if (prevType !== 'single' && type === 'single') {
      resetNoSingle()

      return
    }

    if (type === 'serial' || type === 'parallel' || type === 'custom') {
      if (typeof newNum !== 'number' || newNum < 0) {
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

    if (type === 'suspend') {
      if (prevType === 'serial' || prevType === 'parallel' || prevType === 'custom') {
        resetNoSingle()
      }

      setOtherTypes(type)
    }
  }

  const submit = (template: Template) => {
    if (storeTemplates.some((t) => t.name === template.name)) {
      dispatch(
        setAlert({
          type: 'warning',
          message: T('newW.messages.redundant', intl),
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

  const switchExpand = (index: number) => () => {
    if (index > templates.length) {
      dispatch(
        setAlert({
          type: 'warning',
          message: T('newW.messages.m1', intl),
        })
      )

      return
    }

    setExpand(
      expand === index
        ? (function () {
            dispatch(resetNewExperiment())

            return -1
          })()
        : index
    )
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
          const { conditionalBranches } = values

          const addBranch = (d: Branch) => () => {
            if (d.target === '') {
              dispatch(
                setAlert({
                  type: 'warning',
                  message: T('newW.messages.m2', intl),
                })
              )

              return
            }

            setFieldValue(
              'conditionalBranches',
              conditionalBranches.concat([
                {
                  target: '',
                  expression: '',
                },
              ])
            )
            setNum(num + 1)
          }

          const removeBranch = (index: number) => () => {
            setFieldValue(
              'conditionalBranches',
              conditionalBranches.filter((_: any, i: number) => index !== i)
            )
            setNum(num - 1)
            setTemplates(templates.filter((_: any, i: number) => index !== i))
          }

          const onChange = (index: number) => (_: any, newVal: string | null, reason: string) => {
            const name = `conditionalBranches[${index}].target`

            if (reason === 'clear') {
              setFieldValue(name, '')

              return
            }

            setFieldValue(name, newVal)

            if (templateNames.includes(newVal!)) {
              const template = [...storeTemplates, ...templates].find((t) => t.name === newVal)!

              const tmp = JSON.parse(JSON.stringify(templates))
              tmp[index] = template

              setTemplates(tmp)
              setNum(tmp.length)
            }
          }

          console.log('type of template:' + typeOfTemplate)

          return (
            <>
              <StepLabel icon={<AddCircleIcon color="primary" />}>
                <Space direction="row">
                  <SelectField className={classes.field} name="type" label={T('newW.node.choose')}>
                    {types.map((d) => (
                      <MenuItem key={d} value={d}>
                        <Typography variant="body2">{T(`newW.node.${d}`)}</Typography>
                      </MenuItem>
                    ))}
                  </SelectField>
                  {values.type !== 'custom' && num > 0 && (
                    <TextField
                      className={classes.field}
                      type="number"
                      name="num"
                      label={T('newW.node.number')}
                      inputProps={{ min: 1 }}
                    />
                  )}
                  {update !== undefined && (
                    <Button variant="outlined" startIcon={<CloseIcon />} onClick={updateCallback}>
                      {T('common.cancelEdit')}
                    </Button>
                  )}
                </Space>
              </StepLabel>

              {(values.type === 'serial' || values.type === 'parallel') && (
                <>
                  <SerialOrParallel
                    name={values.name}
                    deadline={values.deadline}
                    type={values.type as TemplateType}
                    childrenCount={values.num}
                    submitTemplate={submit}
                    templates={templates}
                  ></SerialOrParallel>
                </>
              )}
            </>
          )
        }}
      </Formik>

      {isRenderedHTTPTask() && (
        <Box mt={3}>
          <HTTPTask submit={submit} />
        </Box>
      )}
      {num < 0 && (
        <Box ml={8}>
          {typeOfTemplate === 'suspend' && (
            <Box mt={3}>
              <Suspend initialValues={initialValues} submit={submit} />
            </Box>
          )}

          {typeOfTemplate === 'single' && (
            <Box display={otherTypes === 'suspend' ? 'none' : 'initial'}>
              <NewExperimentNext ref={newERef} onSubmit={onSubmit} inWorkflow={true} />
            </Box>
          )}
        </Box>
      )}
    </>
  )
}

export default Add
