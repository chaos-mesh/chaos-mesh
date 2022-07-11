import React from 'react'

import AutocompleteField from '../../esm/AutocompleteField'
import type { AutocompleteFieldProps } from '../../esm/AutocompleteField'

export default {
  title: 'Form/AutoCompleteField',
  component: AutocompleteField,
}

const fruits = ['Apple 🍎', 'Pear 🍐', 'Orange 🍊']

const Template = (props: AutocompleteFieldProps) => {
  return <AutocompleteField {...props} options={fruits} sx={{ width: 320 }} />
}

export const Default = Template.bind({})
Default.args = {}

const fieldInfo = {
  label: 'Fruits',
  helperText: 'Select a fruit',
}

export const LabelAndHelperText = Template.bind({})
LabelAndHelperText.args = {
  ...fieldInfo,
}

export const Multiple = Template.bind({})
Multiple.args = {
  ...fieldInfo,
  helperText: 'Select fruits',
  multiple: true,
}
