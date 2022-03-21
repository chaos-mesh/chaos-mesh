import { ComponentMeta, ComponentStory } from '@storybook/react'
import React, { useState } from 'react'

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
} as ComponentMeta<typeof Checkbox>

const Template: ComponentStory<typeof Checkbox> = ({ ...props }: CheckboxProps) => {
  return <Checkbox {...props} />
}

export const Default = Template.bind({})
Default.args = {
  name: 'spec.abort',
  label: 'Abort HTTP Request',
  helperText: 'Abort is a rule to abort a http session.',
  checked: false,
  disabled: false,
  error: false,
}

export const Disabled = Template.bind({})
Disabled.args = {
  name: 'spec.abort',
  label: 'Abort HTTP Request',
  helperText: 'Abort is a rule to abort a http session.',
  checked: false,
  disabled: true,
  error: false,
}

export const WithValidationError = Template.bind({})
WithValidationError.args = {
  name: 'spec.abort',
  label: 'Abort HTTP Request',
  helperText: 'Abort could not be used with action: delay',
  checked: true,
  disabled: false,
  error: true,
}

export const WithoutHelperText = Template.bind({})
WithoutHelperText.args = {
  name: 'spec.abort',
  label: 'Abort HTTP Request',
  checked: true,
  disabled: false,
  error: true,
}
