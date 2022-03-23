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

import { FormControl, FormControlLabel, FormHelperText, FormLabel } from '@mui/material'
import { default as MuiCheckbox, CheckboxProps as MuiCheckboxProps } from '@mui/material/Checkbox'

export type CheckboxProps = MuiCheckboxProps & {
  /**
   * The label of the checkbox, would be shown on the right of the checkbox.
   */
  label: string
  /**
   * The helper text or error message of the checkbox, would be shown on the bottom of the checkbox in the same line.
   *
   * When field error is true, it should be the error message, otherwise it should be the helper text.
   */
  helperText?: string

  /**
   * Presents there are validation errors for this components.
   */
  error: boolean
}

export default ({ label, helperText, error, ...rest }: CheckboxProps) => {
  return (
    <FormControl error={error} required={true}>
      <FormControlLabel control={<MuiCheckbox {...rest}></MuiCheckbox>} label={label}></FormControlLabel>
      <FormHelperText>{helperText}</FormHelperText>
    </FormControl>
  )
}
