import { MenuItem, SelectChangeEvent } from '@mui/material'
import React, { useState } from 'react'

import SelectField from '../../esm/SelectField'
import type { SelectFieldProps } from '../../esm/SelectField'

export default {
  title: 'Form/SelectField',
  component: SelectField,
}

const fruits = ['🍎 Apple', '🍐 Pear', '🍊 Orange']

const Template = (props: SelectFieldProps) => {
  const [value, setValue] = useState<string>('')

  const onChange = (event: SelectChangeEvent) => {
    setValue(event.target.value)
  }

  return (
    <SelectField {...props} value={value} onChange={onChange} sx={{ width: 320 }}>
      {fruits.map((d) => (
        <MenuItem key={d} value={d}>
          {d}
        </MenuItem>
      ))}
    </SelectField>
  )
}

export const Default = Template.bind({})

const fieldInfo = {
  label: 'Fruits',
  helperText: 'Select a fruit',
}

export const LabelAndHelperText = Template.bind({})
LabelAndHelperText.args = {
  ...fieldInfo,
}

export const Disabled = Template.bind({})
Disabled.args = {
  ...fieldInfo,
  helperText: 'You can not select a fruit for now',
  disabled: true,
}

export const Error = Template.bind({})
Error.args = {
  ...fieldInfo,
  helperText: 'You must select a fruit',
  error: true,
}
