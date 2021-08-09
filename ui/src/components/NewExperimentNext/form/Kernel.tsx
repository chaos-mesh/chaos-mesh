import { Box, IconButton, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, Submit, TextField } from 'components/FormField'
import { useEffect, useState } from 'react'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import Paper from 'components-mui/Paper'
import RemoveCircleIcon from '@material-ui/icons/RemoveCircle'
import Space from 'components-mui/Space'
import targetData from '../data/target'
import { useStoreSelector } from 'store'

interface KernelProps {
  onSubmit: (values: Record<string, any>) => void
}

const Kernel: React.FC<KernelProps> = ({ onSubmit }) => {
  const { target } = useStoreSelector((state) => state.experiments)

  const initialValues = targetData.KernelChaos.spec!

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    setInit({
      fail_kern_request: {
        ...initialValues.fail_kern_request,
        ...target['kernel_chaos']?.fail_kern_request,
      },
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [target])

  return (
    <Formik enableReinitialize initialValues={init} onSubmit={onSubmit}>
      {({ values, setFieldValue }) => {
        const callchain = (values.fail_kern_request as any).callchain

        const addFrame = () =>
          setFieldValue(
            'fail_kern_request.callchain',
            callchain.concat([
              {
                funcname: '',
                parameters: '',
                predicate: '',
              },
            ])
          )

        const removeFrame = (index: number) => () => {
          setFieldValue(
            'fail_kern_request.callchain',
            callchain.filter((_: any, i: number) => index !== i)
          )
        }

        return (
          <Form>
            <Paper sx={{ mb: 6 }}>
              <Box display="flex" justifyContent="space-between" alignItems="center">
                <Typography component="div">Callchain</Typography>
                <IconButton color="primary" size="small" onClick={addFrame}>
                  <AddCircleIcon />
                </IconButton>
              </Box>
              {callchain.length > 0 && (
                <Space mt={6}>
                  {callchain.map((_: any, i: number) => (
                    <Space key={'frame' + i}>
                      <Box display="flex" justifyContent="space-between" alignItems="center">
                        <Typography variant="body2">Frame {i + 1}</Typography>
                        <IconButton color="secondary" size="small" onClick={removeFrame(i)}>
                          <RemoveCircleIcon />
                        </IconButton>
                      </Box>
                      <TextField name={`fail_kern_request.callchain[${i}].funcname`} label="funcname" />
                      <TextField name={`fail_kern_request.callchain[${i}].parameters`} label="parameters" />
                      <TextField name={`fail_kern_request.callchain[${i}].predicate`} label="predicate" />
                    </Space>
                  ))}
                </Space>
              )}
            </Paper>
            <Space>
              <SelectField
                name="fail_kern_request.failtype"
                label="Failtype"
                helperText="What to fail, can be set to 0 / 1 / 2"
              >
                {[0, 1, 2].map((option) => (
                  <MenuItem key={option} value={option}>
                    {option}
                  </MenuItem>
                ))}
              </SelectField>
              <LabelField
                name="fail_kern_request.headers"
                label="Headers"
                helperText="Type string and end with a space to generate the appropriate kernel headers"
              />
              <TextField
                type="number"
                name="fail_kern_request.probability"
                helperText="The fails with probability"
                InputProps={{
                  endAdornment: <InputAdornment position="end">%</InputAdornment>,
                }}
              />
              <TextField type="number" name="fail_kern_request.times" helperText="The max times of failures" />
            </Space>

            <Submit />
          </Form>
        )
      }}
    </Formik>
  )
}

export default Kernel
