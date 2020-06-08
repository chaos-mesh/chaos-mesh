import { SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import { MenuItem } from '@material-ui/core'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { upperFirst } from 'lib/utils'

const actions = ['bandwidth', 'corrupt', 'delay', 'duplicate', 'loss']

export default function NetworkPanel(props: StepperFormTargetProps) {
  const { values, handleChange, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.network_chaos.action"
        name="target.network_chaos.action"
        label="Action"
        helperText="Please select a action"
        value={values.target.network_chaos.action}
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
            label="Buffer"
            helperText="The buffer of bandwidth"
            value={values.target.network_chaos.bandwidth.buffer}
            onChange={handleChange}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.limit"
            label="Limit"
            helperText="The limit of bandwidth"
            value={values.target.network_chaos.bandwidth.limit}
            onChange={handleChange}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.minburst"
            label="Minburst"
            helperText="The minburst of bandwidth"
            value={values.target.network_chaos.bandwidth.minburst}
            onChange={handleChange}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.peakrate"
            label="Peakrate"
            helperText="The peakrate of bandwidth"
            value={values.target.network_chaos.bandwidth.peakrate}
            onChange={handleChange}
          />
          <TextField
            id="target.network_chaos.bandwidth.rate"
            label="Rate"
            helperText="The rate of bandwidth"
            value={values.target.network_chaos.bandwidth.rate}
            onChange={handleChange}
          />
        </>
      )}

      {values.target.network_chaos.action === 'corrupt' && (
        <>
          <TextField
            id="target.network_chaos.corrupt.corrupt"
            label="Corrupt"
            helperText="The corrupt"
            value={values.target.network_chaos.corrupt.corrupt}
            onChange={handleChange}
          />
          <TextField
            id="target.network_chaos.corrupt.correlation"
            label="Correlation"
            helperText="The correlation of corrupt"
            value={values.target.network_chaos.corrupt.correlation}
            onChange={handleChange}
          />
        </>
      )}

      {values.target.network_chaos.action === 'delay' && (
        <>
          <TextField
            id="target.network_chaos.delay.latency"
            label="Latency"
            helperText="The latency of delay"
            value={values.target.network_chaos.delay.latency}
            onChange={handleChange}
          />
          <AdvancedOptions>
            <TextField
              id="target.network_chaos.delay.correlation"
              label="Correlation"
              helperText="The correlation of delay"
              value={values.target.network_chaos.delay.correlation}
              onChange={handleChange}
            />
            <TextField
              id="target.network_chaos.delay.jitter"
              label="Jitter"
              helperText="The jitter of delay"
              value={values.target.network_chaos.delay.jitter}
              onChange={handleChange}
            />
          </AdvancedOptions>
        </>
      )}

      {values.target.network_chaos.action === 'duplicate' && (
        <>
          <TextField
            id="target.network_chaos.duplicate.duplicate"
            label="Duplicate"
            helperText="The duplicate"
            value={values.target.network_chaos.duplicate.duplicate}
            onChange={handleChange}
          />
          <TextField
            id="target.network_chaos.duplicate.correlation"
            label="Correlation"
            helperText="The correlation of duplicate"
            value={values.target.network_chaos.duplicate.correlation}
            onChange={handleChange}
          />
        </>
      )}

      {values.target.network_chaos.action === 'loss' && (
        <>
          <TextField
            id="target.network_chaos.loss.loss"
            label="Loss"
            helperText="The loss"
            value={values.target.network_chaos.loss.loss}
            onChange={handleChange}
          />
          <TextField
            id="target.network_chaos.loss.correlation"
            label="Correlation"
            helperText="The correlation of loss"
            value={values.target.network_chaos.loss.correlation}
            onChange={handleChange}
          />
        </>
      )}
    </>
  )
}
