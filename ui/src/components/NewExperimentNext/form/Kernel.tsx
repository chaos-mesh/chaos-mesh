import { Box, Button, Divider, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, Submit, TextField } from 'components/FormField'
import React, { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import Paper from 'components-mui/Paper'
import RemoveIcon from '@material-ui/icons/Remove'
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
            <Paper>
              <Box display="flex" justifyContent="space-between" alignItems="center">
                <Typography component="div">Callchain</Typography>
                <Button variant="contained" color="primary" startIcon={<AddIcon />} onClick={addFrame}>
                  Add
                </Button>
              </Box>
              {callchain.length > 0 && (
                <Space mt={6}>
                  {callchain.map((_: any, i: number) => (
                    <Box key={'frame' + i}>
                      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
                        <Typography variant="body2" gutterBottom>
                          Frame {i + 1}
                        </Typography>
                        <Button
                          variant="outlined"
                          color="secondary"
                          size="small"
                          startIcon={<RemoveIcon />}
                          onClick={removeFrame(i)}
                        >
                          Remove
                        </Button>
                      </Box>
                      <TextField name={`fail_kern_request.callchain[${i}].funcname`} label="funcname" />
                      <TextField name={`fail_kern_request.callchain[${i}].parameters`} label="parameters" />
                      <TextField name={`fail_kern_request.callchain[${i}].predicate`} label="predicate" />
                    </Box>
                  ))}
                </Space>
              )}
            </Paper>
            <Box my={6}>
              <Divider />
            </Box>
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

            <Submit />
          </Form>
        )
      }}
    </Formik>
  )
}

export default Kernel
