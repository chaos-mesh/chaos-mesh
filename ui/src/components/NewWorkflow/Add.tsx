import { Box, Button, IconButton, MenuItem, StepLabel, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import NewExperimentNext, { NewExperimentHandles } from 'components/NewExperimentNext'
import { SelectField, Submit, TextField } from 'components/FormField'
import { Template, setTemplate, updateTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useEffect, useRef, useState } from 'react'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'
import ArrowRightIcon from '@material-ui/icons/ArrowRight'
import CloseIcon from '@material-ui/icons/Close'
import Custom from './Custom'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import Suspend from './Suspend'
import T from 'components/T'
import _snakecase from 'lodash.snakecase'
import { makeStyles } from '@material-ui/styles'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme) => ({
  field: {
    width: 180,
    marginTop: 0,
    [theme.breakpoints.up('sm')]: {
      marginBottom: 0,
    },
    '& .MuiInputBase-input': {
      padding: 8,
    },
    '& .MuiInputLabel-root, fieldset': {
      fontSize: theme.typography.body2.fontSize,
      lineHeight: 0.875,
    },
  },
}))

const types = ['single', 'serial', 'parallel', 'suspend', 'custom']

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

  const [initialValues, setInitialValues] = useState({
    type: 'single',
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
    ],
  })
  const [num, setNum] = useState(-1)
  const [expand, setExpand] = useState(-1)
  const [otherTypes, setOtherTypes] = useState<'suspend' | 'custom' | ''>('')
  const [templates, setTemplates] = useState<Template[]>([])
  const formRef = useRef<any>()
  const newERef = useRef<NewExperimentHandles>(null)

  const fillExperiment = (t: Template) => {
    const e = t.experiment!

    const kind = e.target.kind

    dispatch(
      setExternalExperiment({
        kindAction: [kind, e.target[_snakecase(kind)].action ?? ''],
        target: e.target,
        basic: e.basic,
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
          const templates = children!

          setTemplates(templates)
          setNum(templates.length)

          break
        case 'suspend':
        case 'custom':
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
    if (type !== 'suspend' && type !== 'custom') {
      setOtherTypes('')
    }

    const prevType = formRef.current.values.type

    if (prevType !== 'single' && type === 'single') {
      resetNoSingle()

      return
    }

    if (type === 'serial' || type === 'parallel') {
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

    if (type === 'suspend' || type === 'custom') {
      if (prevType === 'serial' || prevType === 'parallel') {
        resetNoSingle()
      }

      setOtherTypes(type)
    }
  }

  const submit = (template: Template) => {
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

    const name = experiment.basic.name
    const template = {
      type,
      name,
      experiment,
    }

    submit(template)

    dispatch(resetNewExperiment())
  }

  const submitNoSingleNode = () => {
    const { type, name, deadline } = formRef.current.values
    const template = {
      type,
      name: name.trim(),
      deadline,
      children: templates,
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
        {({ values, errors, touched }) => (
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
                {num > 0 && (
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

            {num > 0 && (
              <Box mt={3} ml={8}>
                <Form>
                  <Paper>
                    <PaperTop title={T(`newW.${values.type}Title`)} boxProps={{ mb: 3 }} />
                    <Space direction="row">
                      <TextField
                        name="name"
                        label={T('common.name')}
                        validate={validateName(T('newW.nameValidation', intl))}
                        helperText={errors.name && touched.name ? errors.name : T('newW.node.nameHelper')}
                        error={errors.name && touched.name ? true : false}
                      />
                      <TextField
                        name="deadline"
                        label={T('newW.node.deadline')}
                        validate={validateDeadline(T('newW.node.deadlineValidation', intl))}
                        helperText={
                          errors.deadline && touched.deadline ? errors.deadline : T('newW.node.deadlineHelper')
                        }
                        error={errors.deadline && touched.deadline ? true : false}
                      />
                    </Space>
                    <Submit disabled={templates.length !== num} />
                  </Paper>
                </Form>

                {Array(num)
                  .fill(0)
                  .map((_, index) => (
                    <Box key={index} ml={8}>
                      <Paper sx={{ my: 6, p: 1.5, borderColor: templates[index] ? 'success.main' : undefined }}>
                        <Box display="flex" alignItems="center">
                          <IconButton size="small" onClick={switchExpand(index)}>
                            {expand === index ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                          </IconButton>
                          <Typography component="div" sx={{ ml: 1 }}>
                            {templates.length > index
                              ? templates[index].name
                              : `${T('newW.node.child', intl)} ${index + 1}`}
                          </Typography>
                        </Box>
                      </Paper>
                      {expand === index && (
                        <Box mt={6}>
                          <Add
                            childIndex={index}
                            parentTemplates={templates}
                            setParentTemplates={setTemplates}
                            setParentExpand={setExpand}
                            externalTemplate={templates[index]}
                          />
                        </Box>
                      )}
                    </Box>
                  ))}
              </Box>
            )}
          </>
        )}
      </Formik>
      {num < 0 && (
        <Box ml={8}>
          <Box display={otherTypes ? 'none' : 'initial'}>
            <NewExperimentNext ref={newERef} onSubmit={onSubmit} inWorkflow={true} />
          </Box>
          {otherTypes === 'suspend' && (
            <Box mt={3}>
              <Suspend initialValues={initialValues} submit={submit} />
            </Box>
          )}
          {otherTypes === 'custom' && (
            <Box mt={3}>
              <Custom initialValues={initialValues} submit={submit} />
            </Box>
          )}
        </Box>
      )}
    </>
  )
}

export default Add
