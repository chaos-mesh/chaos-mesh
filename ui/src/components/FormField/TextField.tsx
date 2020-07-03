import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'

import { Field } from 'formik'
import React from 'react'

const TextField: React.FC<TextFieldProps> = ({ children, ...props }) => (
  <Box mb={2}>
    <Field as={MUITextField} margin="dense" fullWidth variant="outlined" {...props}>
      {children}
    </Field>
  </Box>
)

export default TextField
