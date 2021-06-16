import { FastField, Field, FieldValidator } from 'formik'
import { TextField as MUITextField, TextFieldProps } from '@material-ui/core'

const TextField: React.FC<TextFieldProps & { validate?: FieldValidator; fast?: boolean }> = ({
  fast = false,
  ...rest
}) => {
  const rendered = fast ? (
    <FastField {...rest} as={MUITextField} size="small" fullWidth />
  ) : (
    <Field {...rest} as={MUITextField} size="small" fullWidth />
  )

  return rendered
}

export default TextField
