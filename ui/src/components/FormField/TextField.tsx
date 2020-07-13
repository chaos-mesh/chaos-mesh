import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'
import React, { FC } from 'react'

import { Field } from 'formik'

const TextField: FC<TextFieldProps> = ({ children, fullWidth = true, ...props }) => {
  return (
    <Box mb={2}>
      <Field as={MUITextField} margin="dense" fullWidth={fullWidth} variant="outlined" {...props}>
        {children}
      </Field>
    </Box>
  )
}

export default TextField
