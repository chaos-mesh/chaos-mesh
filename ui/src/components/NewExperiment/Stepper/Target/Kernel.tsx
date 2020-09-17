import { Box, Button, Divider, InputAdornment, MenuItem, Paper, Typography } from '@material-ui/core'
import { LabelField, SelectField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AddIcon from '@material-ui/icons/Add'
import { FormikCtx } from 'components/NewExperiment/types'
import RemoveIcon from '@material-ui/icons/Remove'
import { resetOtherChaos } from 'lib/formikhelpers'
import { useFormikContext } from 'formik'
import PaperTop from 'components/PaperTop'

export default function Kernel() {
  const formikCtx: FormikCtx = useFormikContext()
  const { values, setFieldValue } = formikCtx
  const callchain = values.target.kernel_chaos.fail_kern_request.callchain

  useEffect(() => {
    resetOtherChaos(formikCtx, 'KernelChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const addFrame = () =>
    setFieldValue(
      'target.kernel_chaos.fail_kern_request.callchain',
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
      'target.kernel_chaos.fail_kern_request.callchain',
      callchain.filter((_, i) => index !== i)
    )
  }

  return (
    <>
      <Paper variant="outlined">
        <PaperTop title="Callchain">
          <Button variant="outlined" color="primary" startIcon={<AddIcon />} onClick={addFrame}>
            Add
          </Button>
        </PaperTop>
        <Box>
          {callchain.map((_, i) => (
            <Box key={'frame' + i} p={3}>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
                <Typography variant="body2" gutterBottom>
                  Frame {i + 1}
                </Typography>
                <Button
                  variant="outlined"
                  size="small"
                  color="secondary"
                  startIcon={<RemoveIcon />}
                  onClick={removeFrame(i)}
                >
                  Remove
                </Button>
              </Box>
              <TextField
                id={`target.kernel_chaos.fail_kern_request.callchain[${i}].funcname`}
                name={`target.kernel_chaos.fail_kern_request.callchain[${i}].funcname`}
                label="funcname"
              />
              <TextField
                id={`target.kernel_chaos.fail_kern_request.callchain[${i}].parameters`}
                name={`target.kernel_chaos.fail_kern_request.callchain[${i}].parameters`}
                label="parameters"
              />
              <TextField
                id={`target.kernel_chaos.fail_kern_request.callchain[${i}].predicate`}
                name={`target.kernel_chaos.fail_kern_request.callchain[${i}].predicate`}
                label="predicate"
              />
            </Box>
          ))}
        </Box>
      </Paper>
      <Box my={6}>
        <Divider />
      </Box>
      <SelectField
        id="target.kernel_chaos.fail_kern_request.failtype"
        name="target.kernel_chaos.fail_kern_request.failtype"
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
        id="target.kernel_chaos.fail_kern_request.headers"
        name="target.kernel_chaos.fail_kern_request.headers"
        label="Headers"
        helperText="Type string and end with a space to generate the appropriate kernel headers"
      />
      <TextField
        type="number"
        id="target.kernel_chaos.fail_kern_request.probability"
        name="target.kernel_chaos.fail_kern_request.probability"
        helperText="The fails with probability"
        InputProps={{
          endAdornment: <InputAdornment position="end">%</InputAdornment>,
        }}
      />
      <TextField
        type="number"
        id="target.kernel_chaos.fail_kern_request.times"
        name="target.kernel_chaos.fail_kern_request.times"
        helperText="The max times of failures"
      />
    </>
  )
}
