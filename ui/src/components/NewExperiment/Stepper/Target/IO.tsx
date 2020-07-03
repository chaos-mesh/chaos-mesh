import { InputAdornment, MenuItem } from '@material-ui/core'
import { SelectField, TextField } from 'components/FormField'

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
      >
        {actions.map((option: string) => (
          <MenuItem key={option} value={option}>
            {toTitleCase(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.io_chaos.action !== '' && (
        <AdvancedOptions>
          {values.target.io_chaos.action === ('delay' || 'mixed') && (
            <TextField
              id="target.io_chaos.delay"
              name="target.io_chaos.delay"
              label="Delay"
              helperText="The value of delay action"
            />
          )}

          {values.target.io_chaos.action === ('errno' || 'mixed') && (
            <TextField
              id="target.io_chaos.errno"
              name="target.io_chaos.errno"
              label="Errno"
              helperText="The value of errno action"
            />
          )}
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
            helperText="The path of files for injecting"
          />
          <SelectField
            id="target.io_chaos.methods"
            name="target.io_chaos.methods"
            label="Methods"
            helperText="The IO methods for injecting"
            multiple
          >
            {methods.map((option: string) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </SelectField>
          <TextField
            id="target.io_chaos.addr"
            name="target.io_chaos.addr"
            label="Addr"
            helperText="The sidecar HTTP server address"
          />
        </AdvancedOptions>
      )}
    </>
  )
}
