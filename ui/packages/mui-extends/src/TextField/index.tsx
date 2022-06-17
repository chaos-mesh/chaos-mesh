/*
 * Copyright 2022 Chaos Mesh Authors.
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
import type { TextFieldProps as MuiTextFieldProps, OutlinedInputProps } from '@mui/material'

import FormControl from '../FormControl'
import OutlinedInput from '../OutlinedInput'

export type TextFieldProps = OutlinedInputProps & {
  label?: MuiTextFieldProps['label']
  helperText?: MuiTextFieldProps['helperText']
}

export default function ({ label, helperText, ...rest }: TextFieldProps) {
  const { disabled, error, fullWidth } = rest

  return (
    <FormControl disabled={disabled} error={error} label={label} helperText={helperText} fullWidth={fullWidth}>
      <OutlinedInput {...rest} />
    </FormControl>
  )
}
