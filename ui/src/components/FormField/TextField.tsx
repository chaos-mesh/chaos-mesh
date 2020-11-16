import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'
import { Field, FieldValidator } from 'formik'

import React from 'react'

const TextField: React.FC<TextFieldProps & { validate?: FieldValidator }> = (props) => (
  <Box mb={3}>
    <Field {...props} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  </Box>
)

export default TextField
