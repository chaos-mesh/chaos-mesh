import { Box, Grid, MenuItem, StepLabel, Typography } from '@material-ui/core'
import { Form, Formik, FormikHelpers } from 'formik'
import MultiNode, { MultiNodeHandles } from './MultiNode'
import NewExperimentNext, { NewExperimentHandles } from 'components/NewExperimentNext'
import { SelectField, Submit, TextField } from 'components/FormField'
import { TemplateExperiment, setTemplate } from 'slices/workflows'
import { resetNewExperiment, setExternalExperiment } from 'slices/experiments'
import { useRef, useState } from 'react'
import { validateDuration, validateName } from 'lib/formikhelpers'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import Suspend from './Suspend'
import T from 'components/T'
import _snakecase from 'lodash.snakecase'
import { makeStyles } from '@material-ui/core/styles'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme) => ({
  fields: {
    display: 'flex',
    alignItems: 'center',
    flexWrap: 'wrap',
    [theme.breakpoints.down('xs')]: {
      justifyContent: 'unset',
      '& > *': {
        marginBottom: theme.spacing(3),
        '&:last-child': {
          marginBottom: 0,
        },
      },
    },
  },
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

const types = ['single', 'serial', 'parallel', 'suspend']

const Add = () => {
  const classes = useStyles()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [showNum, setShowNum] = useState(false)
  const [num, setNum] = useState(1)
  const [otherTypes, setOtherTypes] = useState(false)
  const [experiments, setExperiments] = useState<TemplateExperiment[]>([])
  const formRef = useRef<any>()
  const newERef = useRef<NewExperimentHandles>(null)
  const multiNodeRef = useRef<MultiNodeHandles>(null)

  const resetNoSingle = () => {
    setShowNum(false)
    setExperiments([])
    multiNodeRef.current?.setCurrent(0)
  }

  const onValidate = ({ type, num: newNum }: { type: string; num: number }) => {
    if (type !== 'suspend') {
      setOtherTypes(false)
    }

    const prevType = formRef.current.values.type

    if (prevType !== 'single' && type === 'single') {
      resetNoSingle()

      return
    }

    if (type === 'serial' || type === 'parallel') {
      setShowNum(true)

      // Delete extra experiments
      if (num > newNum) {
        setExperiments(experiments.slice(0, -1))
      }

      setNum(newNum)

      return
    }

    if (type === 'suspend') {
      if (prevType === 'serial' || prevType === 'parallel') {
        resetNoSingle()
      }

      setOtherTypes(true)
    }
  }

  const onSubmit = (experiment: any) => {
    const type = formRef.current.values.type

    if (type === 'single') {
      dispatch(
        setTemplate({
          type,
          name: experiment.basic.name,
          experiments: [experiment],
        })
      )
    } else {
      const current = multiNodeRef.current!.current

      multiNodeRef.current!.setCurrent(current + 1)

      // Edit the node that has been submitted before
      if (current < experiments.length) {
        const es = experiments

        es[current] = experiment

        setExperiments(es)

        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
          })
        )
      } else {
        setExperiments([...experiments, experiment])
      }
    }

    newERef.current?.setShowNewPanel('existing')
    dispatch(resetNewExperiment())
  }

  const submitNoSingleNode = (_: any, { resetForm }: FormikHelpers<any>) => {
    const { type, name, duration } = formRef.current.values

    dispatch(
      setTemplate({
        type,
        name,
        duration,
        experiments,
      })
    )

    resetNoSingle()
    resetForm()
  }

  const setCurrentCallback = (index: number) => {
    if (index > experiments.length) {
      dispatch(
        setAlert({
          type: 'warning',
          message: intl.formatMessage({ id: 'newW.messages.m1' }),
        })
      )

      return false
    }

    if (index < experiments.length) {
      const e = experiments[index]

      const kind = e.target.kind

      dispatch(
        setExternalExperiment({
          kindAction: [kind, e.target[_snakecase(kind)].action ?? ''],
          target: e.target,
          basic: e.basic,
        })
      )

      newERef.current?.setShowNewPanel('initial')
    }

    return true
  }

  return (
    <>
      <Formik
        innerRef={formRef}
        initialValues={{ type: 'single', num: 2, name: '', duration: '' }}
        onSubmit={submitNoSingleNode}
        validate={onValidate}
        validateOnBlur={false}
      >
        {({ values, errors, touched }) => (
          <Form>
            <StepLabel icon={<AddCircleIcon color="primary" />}>
              <Space className={classes.fields}>
                <SelectField mb={0} className={classes.field} name="type" label={T('newW.node.choose')}>
                  {types.map((d) => (
                    <MenuItem key={d} value={d}>
                      <Typography variant="body2">{T(`newW.node.${d}`)}</Typography>
                    </MenuItem>
                  ))}
                </SelectField>
                {showNum && (
                  <TextField
                    mb={0}
                    className={classes.field}
                    type="number"
                    name="num"
                    label={T('newW.node.number')}
                    inputProps={{ min: 1 }}
                  />
                )}
              </Space>
            </StepLabel>

            {showNum && (
              <Box mt={6} ml={8}>
                <Paper>
                  <PaperTop title={T(`newW.${values.type}Title`)} />
                  <Grid container spacing={6}>
                    <Grid item xs={12} md={6}>
                      <TextField
                        name="name"
                        label={T('newE.basic.name')}
                        validate={validateName((T('newW.nameValidation') as unknown) as string)}
                        helperText={errors.name && touched.name ? errors.name : T('newW.nameHelper')}
                        error={errors.name && touched.name ? true : false}
                      />
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <TextField
                        name="duration"
                        label={T('newE.schedule.duration')}
                        validate={validateDuration((T('newW.durationValidation') as unknown) as string)}
                        helperText={errors.duration && touched.duration ? errors.duration : T('newW.durationHelper')}
                        error={errors.duration && touched.duration ? true : false}
                      />
                    </Grid>
                  </Grid>
                  <Box display="flex" justifyContent="space-between" alignItems="center">
                    <MultiNode ref={multiNodeRef} count={num} setCurrentCallback={setCurrentCallback} />
                    <Submit mt={0} disabled={experiments.length !== num} />
                  </Box>
                </Paper>
              </Box>
            )}
          </Form>
        )}
      </Formik>
      <Box mt={6} ml={8}>
        <Box style={{ display: otherTypes ? 'none' : 'initial' }}>
          <NewExperimentNext ref={newERef} initPanel="existing" onSubmit={onSubmit} />
        </Box>
        {otherTypes && <Suspend />}
      </Box>
    </>
  )
}

export default Add
