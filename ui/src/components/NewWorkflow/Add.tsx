import { Box, MenuItem, StepLabel, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { SelectField, TextField } from 'components/FormField'
import { useRef, useState } from 'react'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import { Experiment } from 'components/NewExperiment/types'
import NewExperimentNext from 'components/NewExperimentNext'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'
import { setTemplate } from 'slices/workflows'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme) => ({
  field: {
    width: 180,
    marginTop: 0,
    marginBottom: 0,
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
  const formRef = useRef<any>()

  const onValidate = ({ type }: { type: string; num: number }) => {
    setShowNum(type === 'serial' || type === 'parallel')
  }

  const onSubmit = (values: Experiment) => {
    dispatch(
      setTemplate({
        type: formRef.current.values.type,
        experiment: values,
      })
    )

    onSubmitCallback && onSubmitCallback()
  }

  return (
    <>
      <StepLabel icon={<AddCircleIcon color="primary" />}>
        <Formik
          innerRef={formRef}
          initialValues={{ type: 'single', num: 1 }}
          onSubmit={() => {}}
          validate={onValidate}
          validateOnBlur={false}
        >
          <Form>
            <Space display="flex">
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
