import { Box, IconButton, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, Submit, TextField } from 'components/FormField'
import { useEffect, useState } from 'react'

import AddCircleIcon from '@material-ui/icons/AddCircle'
import Paper from 'components-mui/Paper'
import RemoveCircleIcon from '@material-ui/icons/RemoveCircle'
import Space from 'components-mui/Space'
import typesData from '../data/types'
import { useStoreSelector } from 'store'

interface KernelProps {
  onSubmit: (values: Record<string, any>) => void
}

const Kernel: React.FC<KernelProps> = ({ onSubmit }) => {
  const { spec } = useStoreSelector((state) => state.experiments)

  const initialValues = typesData.KernelChaos.spec!

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    setInit({
      failKernRequest: {
        ...initialValues.failKernRequest,
        ...spec.failKernRequest,
      },
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [spec])

  return (
    <Formik enableReinitialize initialValues={init} onSubmit={onSubmit}>
      {({ values, setFieldValue }) => {
        const callchain = (values.failKernRequest as any).callchain

        const addFrame = () =>
          setFieldValue(
            'failKernRequest.callchain',
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
            'failKernRequest.callchain',
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
                      <TextField name={`failKernRequest.callchain[${i}].funcname`} label="funcname" />
                      <TextField name={`failKernRequest.callchain[${i}].parameters`} label="parameters" />
                      <TextField name={`failKernRequest.callchain[${i}].predicate`} label="predicate" />
                    </Space>
                  ))}
                </Space>
              )}
            </Paper>
            <Space>
              <SelectField
                name="failKernRequest.failtype"
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
                name="failKernRequest.headers"
                label="Headers"
                helperText="Type string and end with a space to generate the appropriate kernel headers"
              />
              <TextField
                type="number"
                name="failKernRequest.probability"
                helperText="The fails with probability"
                InputProps={{
                  endAdornment: <InputAdornment position="end">%</InputAdornment>,
                }}
              />
              <TextField type="number" name="failKernRequest.times" helperText="The max times of failures" />
            </Space>

            <Submit />
          </Form>
        )
      }}
    </Formik>
  )
}

export default Kernel
