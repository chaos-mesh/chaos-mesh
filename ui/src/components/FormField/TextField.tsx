import { Box, TextField as MUITextField, TextFieldProps } from '@material-ui/core'
import { FastField, Field, FieldValidator } from 'formik'

import React from 'react'

const TextField: React.FC<TextFieldProps & { validate?: FieldValidator; fast?: boolean; mb?: number }> = ({
  fast = false,
  mb = 1.5,
  ...rest
}) => {
  const rendered = fast ? (
    <FastField {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  ) : (
    <Field {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  )

  return mb > 0 ? <Box mb={mb}>{rendered}</Box> : rendered
}

export default TextField
