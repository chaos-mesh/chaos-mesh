import React, { FC } from 'react'
import { Box, TextField, TextFieldProps } from '@material-ui/core'

const TextInput: FC<TextFieldProps> = ({ children, fullWidth = true, ...inputProps }) => {
  return (
    <Box mb={4}>
      <TextField fullWidth={fullWidth} {...inputProps}>
        {children}
      </TextField>
    </Box>
  )
}

export default TextInput
