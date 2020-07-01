import { SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import { MenuItem } from '@material-ui/core'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { upperFirst } from 'lib/utils'

const actions = ['bandwidth', 'corrupt', 'delay', 'duplicate', 'loss']

export default function NetworkPanel(props: StepperFormTargetProps) {
  const { values, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.network_chaos.action"
        name="target.network_chaos.action"
        label="Action"
        helperText="Please select an action"
        onChange={handleActionChange}
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {upperFirst(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.network_chaos.action === 'bandwidth' && (
        <>
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.buffer"
            name="target.network_chaos.bandwidth.buffer"
            label="Buffer"
            helperText="The buffer of bandwidth"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.limit"
            name="target.network_chaos.bandwidth.limit"
            label="Limit"
            helperText="The limit of bandwidth"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.minburst"
            name="target.network_chaos.bandwidth.minburst"
            label="Minburst"
            helperText="The minburst of bandwidth"
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.peakrate"
            name="target.network_chaos.bandwidth.peakrate"
            label="Peakrate"
            helperText="The peakrate of bandwidth"
          />
          <TextField
            id="target.network_chaos.bandwidth.rate"
            name="target.network_chaos.bandwidth.rate"
            label="Rate"
            helperText="The rate of bandwidth"
          />
        </>
      )}

      {values.target.network_chaos.action === 'corrupt' && (
        <>
          <TextField
            id="target.network_chaos.corrupt.corrupt"
            name="target.network_chaos.corrupt.corrupt"
            label="Corrupt"
            helperText="The corrupt"
          />
          <TextField
            id="target.network_chaos.corrupt.correlation"
            name="target.network_chaos.corrupt.correlation"
            label="Correlation"
            helperText="The correlation of corrupt"
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
              id="target.network_chaos.delay.correlation"
              name="target.network_chaos.delay.correlation"
              label="Correlation"
              helperText="The correlation of delay"
            />
            <TextField
              id="target.network_chaos.delay.jitter"
              name="target.network_chaos.delay.jitter"
              label="Jitter"
              helperText="The jitter of delay"
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
            helperText="The duplicate"
          />
          <TextField
            id="target.network_chaos.duplicate.correlation"
            name="target.network_chaos.duplicate.correlation"
            label="Correlation"
            helperText="The correlation of duplicate"
          />
        </>
      )}

      {values.target.network_chaos.action === 'loss' && (
        <>
          <TextField
            id="target.network_chaos.loss.loss"
            name="target.network_chaos.loss.loss"
            label="Loss"
            helperText="The loss"
          />
          <TextField
            id="target.network_chaos.loss.correlation"
            name="target.network_chaos.loss.correlation"
            label="Correlation"
            helperText="The correlation of loss"
          />
        </>
      )}
    </>
  )
}
