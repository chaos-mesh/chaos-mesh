import React from 'react'

import Checkbox from '../../esm/Checkbox'
import type { CheckboxProps } from '../../esm/Checkbox'

export default {
  title: 'Form/Checkbox',
  component: Checkbox,
  argTypes: {
    onChange: {
      action: 'onChange',
    },
  },
}

const Template = (props: CheckboxProps) => <Checkbox {...props} />

const fieldInfo = {
  name: 'spec.abort',
  label: 'Abort HTTP Request',
  helperText: 'Abort is a rule to abort a http session.',
}

export const Default = Template.bind({})
Default.args = {
  ...fieldInfo,
}

export const Disabled = Template.bind({})
Disabled.args = {
  ...fieldInfo,
  disabled: true,
}

export const WithValidationError = Template.bind({})
WithValidationError.args = {
  ...fieldInfo,
  helperText: 'Abort could not be used with action: delay',
  checked: true,
  error: true,
}

export const WithoutHelperText = Template.bind({})
WithoutHelperText.args = {
  ...fieldInfo,
  helperText: undefined,
  checked: true,
}
