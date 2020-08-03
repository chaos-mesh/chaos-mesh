import { Box, TextField as MUITextField, MenuItem, Typography } from '@material-ui/core'
import { LabelField, TextField } from 'components/FormField'
import React, { useEffect, useRef, useState } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { defaultExperimentSchema } from 'components/NewExperiment/constants'
import { getIn } from 'formik'
import { resetOtherChaos } from 'lib/formikhelpers'

const actions = ['CPU', 'Memory', 'Mixed']

export default function Stress(props: StepperFormTargetProps) {
  const { values, setFieldValue } = props

  const actionRef = useRef('')
  const [action, _setAction] = useState('')
  const setAction = (newVal: string) => {
    actionRef.current = newVal
    _setAction(newVal)
  }

  useEffect(() => {
    resetOtherChaos(props, 'StressChaos', false)

    if (getIn(values, 'target.stress_chaos.stressors.cpu') === null) {
      setFieldValue('target.stress_chaos.stressors.cpu', defaultExperimentSchema.target.stress_chaos.stressors.cpu)
    }

    if (getIn(values, 'target.stress_chaos.stressors.memory') === null) {
      setFieldValue(
        'target.stress_chaos.stressors.memory',
        defaultExperimentSchema.target.stress_chaos.stressors.memory
      )
    }

    // Remove another when choosing a single action
    return () => {
      if (actionRef.current === 'CPU') {
        // Because LabelField will set value when before unmount, it's needed to wrap setFieldValue into setTimeout
        setTimeout(() => setFieldValue('target.stress_chaos.stressors.memory', null))
      } else if (actionRef.current === 'Memory') {
        setTimeout(() => setFieldValue('target.stress_chaos.stressors.cpu', null))
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleActionChange = (e: React.ChangeEvent<HTMLInputElement>) => setAction(e.target.value)

  return (
    <>
      <Box mb={2}>
        <MUITextField
          variant="outlined"
          select
          margin="dense"
          fullWidth
          label="Action"
          helperText="Please select an action"
          value={action}
          onChange={handleActionChange}
        >
          {actions.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </MUITextField>
      </Box>

      {(action === 'CPU' || action === 'Mixed') && (
        <>
          <Box ml={1}>
            <Typography gutterBottom>CPU</Typography>
          </Box>
          <TextField
            type="number"
            id="target.stress_chaos.stressors.cpu.workers"
            name="target.stress_chaos.stressors.cpu.workers"
            label="Workers"
            helperText="CPU workers"
          />
          <TextField
            type="number"
            id="target.stress_chaos.stressors.cpu.load"
            name="target.stress_chaos.stressors.cpu.load"
            label="Load"
            helperText="CPU load"
          />
          <LabelField
            id="target.stress_chaos.stressors.cpu.options"
            name="target.stress_chaos.stressors.cpu.options"
            label="Options of CPU stressors"
            helperText="Type string and end with a space to generate the stress-ng options"
          />
        </>
      )}

      {(action === 'Memory' || action === 'Mixed') && (
        <>
          <Box ml={1}>
            <Typography gutterBottom>Memory</Typography>
          </Box>
          <TextField
            type="number"
            id="target.stress_chaos.stressors.memory.workers"
            name="target.stress_chaos.stressors.memory.workers"
            label="Workers"
            helperText="Memory workers"
          />
          <TextField
            id="target.stress_chaos.stressors.memory.size"
            name="target.stress_chaos.stressors.memory.size"
            label="Size"
            helperText="Memory size"
          />
          <LabelField
            id="target.stress_chaos.stressors.memory.options"
            name="target.stress_chaos.stressors.memory.options"
            label="Options of Memory stressors"
            helperText="Type string and end with a space to generate the stress-ng options"
          />
        </>
      )}

      {action !== '' && (
        <AdvancedOptions>
          <TextField
            id="target.stress_chaos.stressng_stressors"
            name="target.stress_chaos.stressng_stressors"
            label="Options of stress-ng"
            helperText="The options of stress-ng, treated as a string"
          />
        </AdvancedOptions>
      )}
    </>
  )
}
