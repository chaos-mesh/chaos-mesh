import { Box, Button, MenuItem, StepLabel, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import NewExperimentNext, { NewExperimentHandles } from 'components/NewExperimentNext'
import { SelectField, TextField } from 'components/FormField'
import { TemplateExperiment, setTemplate } from 'slices/workflows'
import { useRef, useState } from 'react'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import MultiNode from './MultiNode'
import PublishIcon from '@material-ui/icons/Publish'
import Space from 'components-mui/Space'
import T from 'components/T'
import _snakecase from 'lodash.snakecase'
import { makeStyles } from '@material-ui/core/styles'
import { setAlert } from 'slices/globalStatus'
import { setExternalExperiment } from 'slices/experiments'
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

interface AddProps {
  onSubmitCallback?: () => void
}

const Add: React.FC<AddProps> = ({ onSubmitCallback }) => {
  const classes = useStyles()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [showNum, setShowNum] = useState(false)
  const [num, setNum] = useState(1)
  const [otherTypes, setOtherTypes] = useState(false)
  const [experiments, setExperiments] = useState<TemplateExperiment[]>([])
  const [current, setCurrent] = useState(0)
  const formRef = useRef<any>()
  const newERef = useRef<NewExperimentHandles>(null)

  const resetNoSingle = () => {
    setShowNum(false)
    setExperiments([])
    setCurrent(0)
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
      setCurrent(current + 1)

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

    onSubmitCallback && onSubmitCallback()
  }

  const submitNoSingleNode = () => {
    const { type, name } = formRef.current.values

    dispatch(
      setTemplate({
        type,
        name,
        experiments,
      })
    )

    resetNoSingle()
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
      <StepLabel icon={<AddCircleIcon color="primary" />}>
        <Formik
          innerRef={formRef}
          initialValues={{ type: 'single', num: 2, name: '' }}
          onSubmit={() => {}}
          validate={onValidate}
          validateOnBlur={false}
        >
          <Form>
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
            {showNum && (
              <Box display="flex" justifyContent="space-between" alignItems="center" mt={3}>
                <TextField mb={0} className={classes.field} name="name" label={T('newE.basic.name')} />
                <MultiNode
                  count={num}
                  current={current}
                  setCurrent={setCurrent}
                  setCurrentCallback={setCurrentCallback}
                />
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={<PublishIcon />}
                  disabled={experiments.length !== num}
                  onClick={submitNoSingleNode}
                >
                  {T('common.submit')}
                </Button>
              </Box>
            )}
          </Form>
        </Formik>
      </StepLabel>
      <Box mt={6} ml={8}>
        <Box style={{ display: otherTypes ? 'none' : 'initial' }}>
          <NewExperimentNext ref={newERef} initPanel="existing" onSubmit={onSubmit} />
        </Box>
      </Box>
    </>
  )
}

export default Add
