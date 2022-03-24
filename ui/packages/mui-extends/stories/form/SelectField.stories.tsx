import React, { useState } from 'react'

import { MenuItem } from '@mui/material'
import SelectField from '../../esm/SelectField'
import type { SelectFieldProps } from '../../esm/SelectField'

export default {
  title: 'Form/SelectField',
  component: SelectField,
}

const fruits = ['🍎', '🍐', '🍊']

const Template = ({ data, ...props }: SelectFieldProps & { data: string[] }) => {
  const [value, setValue] = useState(data[0])

  const onChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValue(event.target.value)
  }

  return (
    <SelectField
      fullWidth={false}
      {...props}
      value={value}
      onChange={onChange}
      children={data.map((d) => (
        <MenuItem key={d} value={d}>
          {d}
        </MenuItem>
      ))}
    />
  )
}

export const Default = Template.bind({})
Default.args = { data: fruits }
