import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'

import { Field } from 'formik'
import React from 'react'

const TextField: React.FC<TextFieldProps> = (props) => (
  <Box mb={2}>
    <Field {...props} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  </Box>
)

export default TextField
