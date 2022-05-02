import React from 'react'
import TextField from '../../esm/TextField'
import type { TextFieldProps } from '../../esm/TextField'

export default {
  title: 'Form/TextField',
  component: TextField,
}

const Template = (args: TextFieldProps) => <TextField fullWidth={false} {...args} />

export const Default = Template.bind({})
Default.args = {
  label: 'TextField',
  helperText: 'This is a TextField',
}
