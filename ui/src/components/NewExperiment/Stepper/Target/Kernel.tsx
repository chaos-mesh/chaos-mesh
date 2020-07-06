import { Box, Button, Divider, InputAdornment, MenuItem, Typography } from '@material-ui/core'
import { LabelField, SelectField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'

import AddIcon from '@material-ui/icons/Add'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { resetOtherChaos } from 'lib/formikhelpers'

export default function Kernel(props: StepperFormTargetProps) {
  const { values, setFieldValue } = props
  const callchain = values.target.kernel_chaos.fail_kernel_req.callchain

  useEffect(() => {
    resetOtherChaos(props, 'KernelChaos', false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const addFrame = () =>
    setFieldValue(
      'target.kernel_chaos.fail_kernel_req.callchain',
      callchain.concat([
        {
          funcname: '',
          parameters: '',
          predicate: '',
        },
      ])
    )

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography>Callchain</Typography>
        <Button variant="outlined" color="primary" size="small" startIcon={<AddIcon />} onClick={addFrame}>
          Add
        </Button>
      </Box>
      <Box>
        {callchain.map((frame, i) => (
          <Box key={i}>
            <Typography variant="body2" gutterBottom>
              Frame {i + 1}
            </Typography>
            <TextField
              id={`target.kernel_chaos.fail_kernel_req.callchain[${i}].funcname`}
              name={`target.kernel_chaos.fail_kernel_req.callchain[${i}].funcname`}
              label="funcname"
            />
            <TextField
              id={`target.kernel_chaos.fail_kernel_req.callchain[${i}].parameters`}
              name={`target.kernel_chaos.fail_kernel_req.callchain[${i}].parameters`}
              label="parameters"
            />
            <TextField
              id={`target.kernel_chaos.fail_kernel_req.callchain[${i}].predicate`}
              name={`target.kernel_chaos.fail_kernel_req.callchain[${i}].predicate`}
              label="predicate"
            />
          </Box>
        ))}
      </Box>
      <Box mb={3}>
        <Divider />
      </Box>
      <SelectField
        id="target.kernel_chaos.fail_kernel_req.failtype"
        name="target.kernel_chaos.fail_kernel_req.failtype"
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
        id="target.kernel_chaos.fail_kernel_req.headers"
        name="target.kernel_chaos.fail_kernel_req.headers"
        label="Headers"
        helperText="Type string and end with a space to generate the appropriate kernel headers"
      />
      <TextField
        type="number"
        id="target.kernel_chaos.fail_kernel_req.probability"
        name="target.kernel_chaos.fail_kernel_req.probability"
        helperText="The fails with probability"
        InputProps={{
          endAdornment: <InputAdornment position="end">%</InputAdornment>,
        }}
      />
      <TextField
        type="number"
        id="target.kernel_chaos.fail_kernel_req.times"
        name="target.kernel_chaos.fail_kernel_req.times"
        helperText="The max times of fails"
      />
    </>
  )
}
