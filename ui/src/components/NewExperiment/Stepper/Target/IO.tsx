import { AutocompleteMultipleField, SelectField, TextField } from 'components/FormField'
import { InputAdornment, MenuItem } from '@material-ui/core'

import AdvancedOptions from 'components/AdvancedOptions'
import React from 'react'
import { StepperFormTargetProps } from 'components/NewExperiment/types'
import { toTitleCase } from 'lib/utils'

const actions = ['delay', 'errno', 'mixed']
const methods = [
  'open',
  'read',
  'write',
  'mkdir',
  'rmdir',
  'opendir',
  'fsync',
  'flush',
  'release',
  'truncate',
  'getattr',
  'chown',
  'chmod',
  'utimens',
  'allocate',
  'getlk',
  'setlk',
  'setlkw',
  'statfs',
  'readlink',
  'symlink',
  'create',
  'access',
  'link',
  'mknod',
  'rename',
  'unlink',
  'getxattr',
  'listxattr',
  'removexattr',
  'setxattr',
]

export default function IO(props: StepperFormTargetProps) {
  const { values, handleActionChange } = props

  return (
    <>
      <SelectField
        id="target.io_chaos.action"
        name="target.io_chaos.action"
        label="Action"
        helperText="Please select an action"
        onChange={handleActionChange}
        onBlur={() => {}} // Delay the form validation with an empty func. If don’t do this, errors will appear early
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {toTitleCase(option)}
          </MenuItem>
        ))}
      </SelectField>

      {(values.target.io_chaos.action === 'delay' || values.target.io_chaos.action === 'mixed') && (
        <TextField
          id="target.io_chaos.delay"
          name="target.io_chaos.delay"
          label="Delay"
          helperText="Optional. The value of delay of I/O operations. If it's empty, the operator will generate a value for it randomly."
        />
      )}

      {(values.target.io_chaos.action === 'errno' || values.target.io_chaos.action === 'mixed') && (
        <TextField
          id="target.io_chaos.errno"
          name="target.io_chaos.errno"
          label="Errno"
          helperText="Optional. The error code returned by I/O operators. By default, it returns a random error code"
        />
      )}

      {values.target.io_chaos.action !== '' && (
        <AdvancedOptions>
          <TextField
            id="target.io_chaos.percent"
            name="target.io_chaos.percent"
            label="Percent"
            helperText="The percentage of injection errors"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
          />
          <TextField
            id="target.io_chaos.path"
            name="target.io_chaos.path"
            label="Path"
            helperText="The path of files for injecting. If it's empty, the IOChaos action will inject into all files."
          />
          <AutocompleteMultipleField
            id="target.io_chaos.methods"
            name="target.io_chaos.methods"
            label="Methods"
            helperText="Optional. The IO methods for injecting IOChaos actions"
            options={methods}
          />
          <TextField
            id="target.io_chaos.addr"
            name="target.io_chaos.addr"
            label="Addr"
            helperText="Optional. The sidecar HTTP server address. By default, it will be set to :65534"
          />
        </AdvancedOptions>
      )}
    </>
  )
}
