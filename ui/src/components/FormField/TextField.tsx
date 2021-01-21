import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'
import { FastField, Field, FieldValidator } from 'formik'

import React from 'react'

const TextField: React.FC<TextFieldProps & { validate?: FieldValidator; fast?: boolean }> = ({
  fast = false,
  ...rest
}) => (
  <Box mb={3}>
    {fast ? (
      <FastField {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
    ) : (
      <Field {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
    )}
  </Box>
)

export default TextField
