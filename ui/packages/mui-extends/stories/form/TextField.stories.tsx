import HelpOutlineIcon from '@mui/icons-material/HelpOutline'
import SearchIcon from '@mui/icons-material/Search'
import { InputAdornment } from '@mui/material'
import React from 'react'

import Space from '../../esm/Space'
import TextField from '../../esm/TextField'
import type { TextFieldProps } from '../../esm/TextField'

export default {
  title: 'Form/TextField',
  component: TextField,
}

const Template = (args: TextFieldProps) => <TextField {...args} sx={{ width: 320 }} />

export const Default = Template.bind({})
Default.args = {
  placeholder: 'Type something...',
}

const fieldInfo = {
  placeholder: '@every 30s',
  label: 'Schedule',
  helperText: 'if u dont know type what...',
}

export const LabelAndHelperText = Template.bind({})
LabelAndHelperText.args = {
  ...fieldInfo,
}

export const HasInputAdornment = Template.bind({})
HasInputAdornment.args = {
  startAdornment: (
    <InputAdornment position="start">
      <SearchIcon fontSize="small" />
    </InputAdornment>
  ),
  endAdornment: (
    <InputAdornment position="end">
      <HelpOutlineIcon fontSize="small" />
    </InputAdornment>
  ),
}

export const Group = () => (
  <Space>
    <Template label="Name" placeholder="experiment-1" helperText="This field is required" />
    <Template label="Namespace" placeholder="chaos-mesh" helperText="This field is optional" />
  </Space>
)

export const Disabled = Template.bind({})
Disabled.args = {
  ...fieldInfo,
  disabled: true,
}

export const Error = Template.bind({})
Error.args = {
  ...fieldInfo,
  error: true,
}
