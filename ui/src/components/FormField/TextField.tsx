/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
