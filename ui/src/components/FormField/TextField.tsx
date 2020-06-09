import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'
import React, { FC } from 'react'

const TextField: FC<TextFieldProps> = ({ children, fullWidth = true, ...props }) => {
  return (
    <Box mb={2}>
      <MUITextField margin="dense" fullWidth={fullWidth} variant="outlined" {...props}>
        {children}
      </MUITextField>
    </Box>
  )
}

export default TextField
