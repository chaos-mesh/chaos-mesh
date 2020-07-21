import { InputAdornment, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { toTitleCase } from 'lib/utils'

const actions = ['loss', 'delay', 'duplicate', 'corrupt', 'bandwidth']

export default function Network(props: StepperFormTargetProps) {
  const { values, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.network_chaos.action"
        name="target.network_chaos.action"
        label="Action"
        helperText="Please select a NetworkChaos action"
        onChange={handleActionChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {toTitleCase(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.network_chaos.action === 'bandwidth' && (
        <>
          <TextField
            id="target.network_chaos.bandwidth.rate"
            name="target.network_chaos.bandwidth.rate"
            label="Rate"
            helperText="The rate allows bps, kbps, mbps, gbps, tbps unit. For example, bps means bytes per second"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.buffer"
            name="target.network_chaos.bandwidth.buffer"
            label="Buffer"
            helperText="The maximum amount of bytes that tokens can be available instantaneously"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.limit"
            name="target.network_chaos.bandwidth.limit"
            label="Limit"
            helperText="The number of bytes that can be queued waiting for tokens to become available"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.peakrate"
            name="target.network_chaos.bandwidth.peakrate"
            label="Peakrate"
            helperText="The maximum depletion rate of the bucket"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.minburst"
            name="target.network_chaos.bandwidth.minburst"
            label="Minburst"
            helperText="The size of the peakrate bucket"
          />
        </>
      )}

      {values.target.network_chaos.action === 'corrupt' && (
        <>
          <TextField
            id="target.network_chaos.corrupt.corrupt"
            name="target.network_chaos.corrupt.corrupt"
            label="Corrupt"
            helperText="The percentage of packet corruption"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
          <TextField
            id="target.network_chaos.corrupt.correlation"
            name="target.network_chaos.corrupt.correlation"
            label="Correlation"
            helperText="The correlation of corrupt"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </>
      )}

      {values.target.network_chaos.action === 'delay' && (
        <>
          <TextField
            id="target.network_chaos.delay.latency"
            name="target.network_chaos.delay.latency"
            label="Latency"
            helperText="The latency of delay"
          />
          <AdvancedOptions>
            <TextField
              id="target.network_chaos.delay.jitter"
              name="target.network_chaos.delay.jitter"
              label="Jitter"
              helperText="The jitter of delay"
            />
            <TextField
              id="target.network_chaos.delay.correlation"
              name="target.network_chaos.delay.correlation"
              label="Correlation"
              helperText="The correlation of delay"
              InputProps={{
                endAdornment: <InputAdornment position="end">%</InputAdornment>,
              }}
            />
          </AdvancedOptions>
        </>
      )}

      {values.target.network_chaos.action === 'duplicate' && (
        <>
          <TextField
            id="target.network_chaos.duplicate.duplicate"
            name="target.network_chaos.duplicate.duplicate"
            label="Duplicate"
            helperText="The percentage of packet duplication"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
          <TextField
            id="target.network_chaos.duplicate.correlation"
            name="target.network_chaos.duplicate.correlation"
            label="Correlation"
            helperText="The correlation of duplicate"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </>
      )}

      {values.target.network_chaos.action === 'loss' && (
        <>
          <TextField
            id="target.network_chaos.loss.loss"
            name="target.network_chaos.loss.loss"
            label="Loss"
            helperText="The percentage of packet loss"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
          <TextField
            id="target.network_chaos.loss.correlation"
            name="target.network_chaos.loss.correlation"
            label="Correlation"
            helperText="The correlation of loss"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
        </>
      )}
    </>
  )
}
