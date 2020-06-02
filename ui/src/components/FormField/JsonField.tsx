import React from 'react'
import TextField from './TextField'
import { TextFieldProps } from '@material-ui/core'

interface JsonFieldProps {
  onChangeCallback: (e: React.ChangeEvent<HTMLInputElement>, formatted: string) => void
}

const JsonField: React.FC<TextFieldProps & JsonFieldProps> = ({ onChangeCallback, ...props }) => {
  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value

    try {
      const obj = JSON.parse(val)

      onChangeCallback(e, JSON.stringify(obj, null, 4))
    } catch {
      onChangeCallback(e, val)
    }
  }

  return <TextField onChange={onChange} {...props} />
}

export default JsonField
