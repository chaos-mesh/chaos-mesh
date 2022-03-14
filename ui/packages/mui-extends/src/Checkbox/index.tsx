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
   * The name of ths input field when using it in a form.
   */
  name?: string
  /**
   * The label of the checkbox, would be shown on the top of the checkbox.
   */
  label: string
  /**
   * The helper text of the checkbox, would be shown on the right of the checkbox in the same line.
   *
   * If not provided, it would use the `label` as the helper text.
   */
  helperText?: string
  /**
   * Validation error message, would be shown on the bottom of the checkbox.
   */
  errorMessage?: string
}

export default (props: CheckboxProps) => {
  return (
    <div>
      <FormControl error={props.errorMessage !== ''} required={true}>
        <FormLabel component="legend">{props.label}</FormLabel>
        <FormControlLabel
          control={
            <MuiCheckbox
              checked={props.checked}
              onChange={props.onChange}
              disabled={props.disabled}
              inputProps={{ name: props.name }}
            ></MuiCheckbox>
          }
          label={props.helperText || props.label}
        ></FormControlLabel>

        <FormHelperText>{props.errorMessage}</FormHelperText>
      </FormControl>
    </div>
  )
}
