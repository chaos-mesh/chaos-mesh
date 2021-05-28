import { FastField, Field, FieldValidator } from 'formik'
import { TextField as MUITextField, TextFieldProps } from '@material-ui/core'

import React from 'react'

const TextField: React.FC<TextFieldProps & { validate?: FieldValidator; fast?: boolean }> = ({
  fast = false,
  ...rest
}) => {
  const rendered = fast ? (
    <FastField {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  ) : (
    <Field {...rest} as={MUITextField} variant="outlined" margin="dense" fullWidth />
  )

  return rendered
}

export default TextField
