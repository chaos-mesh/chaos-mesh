import { AutocompleteMultipleField, LabelField, SelectField, TextField } from 'components/FormField'
import { FormikCtx, StepperFormTargetProps } from 'components/NewExperiment/types'
import { InputAdornment, MenuItem } from '@material-ui/core'

import AdvancedOptions from 'components/AdvancedOptions'
import React from 'react'
import { toTitleCase } from 'lib/utils'
import { useFormikContext } from 'formik'

const actions = ['latency', 'fault', 'attrOverride']
const methods = [
  'lookup',
  'forget',
  'getattr',
  'setattr',
  'readlink',
  'mknod',
  'mkdir',
  'unlink',
  'rmdir',
  'symlink',
  'rename',
  'link',
  'open',
  'read',
  'write',
  'flush',
  'release',
  'fsync',
  'opendir',
  'readdir',
  'releasedir',
  'fsyncdir',
  'statfs',
  'setxattr',
  'getxattr',
  'listxattr',
  'removexattr',
  'access',
  'create',
  'getlk',
  'setlk',
  'bmap',
]

export default function IO(props: StepperFormTargetProps) {
  const { values, errors, touched }: FormikCtx = useFormikContext()
  const { handleActionChange } = props

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

      {values.target.io_chaos.action !== '' && (
        <>
          {values.target.io_chaos.action === 'latency' && (
            <TextField
              id="target.io_chaos.delay"
              name="target.io_chaos.delay"
              label="Delay"
              helperText="The value of delay of I/O operations. If it's empty, the operator will generate a value for it randomly."
              error={errors.target?.io_chaos?.delay && touched.target?.io_chaos?.delay ? true : false}
            />
          )}

          {values.target.io_chaos.action === 'fault' && (
            <TextField
              type="number"
              inputProps={{ min: 0 }}
              id="target.io_chaos.errno"
              name="target.io_chaos.errno"
              label="Errno"
              helperText="The error code returned by I/O operators. By default, it returns a random error code"
              error={errors.target?.io_chaos?.errno && touched.target?.io_chaos?.errno ? true : false}
            />
          )}

          {values.target.io_chaos.action === 'attrOverride' && (
            <LabelField
              id="target.io_chaos.attr"
              name="target.io_chaos.attr"
              label="Attr"
              isKV
              error={errors.target?.io_chaos?.attr && touched.target?.io_chaos?.attr ? true : false}
            />
          )}

          <TextField
            id="target.io_chaos.volume_path"
            name="target.io_chaos.volume_path"
            label="Volume Path"
            helperText="The mount path of injected volume"
          />

          <AdvancedOptions>
            <TextField
              type="number"
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
              helperText="Optional. The path of files for injecting. If it's empty, the action will inject into all files."
            />
            <AutocompleteMultipleField
              id="target.io_chaos.methods"
              name="target.io_chaos.methods"
              label="Methods"
              helperText="Optional. The IO methods for injecting IOChaos actions"
              options={methods}
            />
          </AdvancedOptions>
        </>
      )}
    </>
  )
}
