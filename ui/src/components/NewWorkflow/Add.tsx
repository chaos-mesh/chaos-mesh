import { Box, Button, MenuItem, StepLabel, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { SelectField, TextField } from 'components/FormField'
import { TemplateExperiment, setTemplate } from 'slices/workflows'
import { useRef, useState } from 'react'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import MultiNode from './MultiNode'
import NewExperimentNext from 'components/NewExperimentNext'
import PublishIcon from '@material-ui/icons/Publish'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'
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

const types = ['single', 'serial', 'parallel']

interface AddProps {
  onSubmitCallback?: () => void
}

const Add: React.FC<AddProps> = ({ onSubmitCallback }) => {
  const classes = useStyles()

  const dispatch = useStoreDispatch()

  const [showNum, setShowNum] = useState(false)
  const [num, setNum] = useState(1)
  const [experiments, setExperiments] = useState<TemplateExperiment[]>([])
  const [current, setCurrent] = useState(0)
  const formRef = useRef<any>()

  const resetNoSingle = () => {
    setShowNum(false)
    setExperiments([])
    setCurrent(0)
  }

  const onValidate = ({ type, num: newNum }: { type: string; num: number }) => {
    if (formRef.current.values.type !== 'single' && type === 'single') {
      resetNoSingle()
    }

    if (type === 'serial' || type === 'parallel') {
      setShowNum(true)
    }

    // Delete extra experiments
    if (num > newNum) {
      setExperiments(experiments.slice(0, -1))
    }

    setNum(newNum)
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
      setExperiments([...experiments, experiment])
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
              <Space display="flex" justifyContent="space-between" alignItems="center" mt={3}>
                <TextField mb={0} className={classes.field} name="name" label={T('newE.basic.name')} />
                <MultiNode count={num} current={current} setCurrent={setCurrent} />
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={<PublishIcon />}
                  disabled={experiments.length !== num}
                  onClick={submitNoSingleNode}
                >
                  {T('common.submit')}
                </Button>
              </Space>
            )}
          </Form>
        </Formik>
      </StepLabel>
      <Box mt={6} ml={8}>
        <NewExperimentNext initPanel="existing" onSubmit={onSubmit} />
      </Box>
    </>
  )
}

export default Add
