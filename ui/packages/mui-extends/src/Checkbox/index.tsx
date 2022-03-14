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

import { FormControl, FormControlLabel, FormHelperText, FormLabel, Input, InputLabel } from '@mui/material'
import { default as MuiCheckbox, CheckboxProps as MuiCheckboxProps } from '@mui/material/Checkbox'

export type CheckboxProps = MuiCheckboxProps & {
  name: string
  label: string
  helperText: string
  errorMessage: string
}

export default (props: CheckboxProps) => {
  return (
    <div>
      <FormControl error={props.errorMessage !== ''} required={true}>
        <FormLabel component="legend">{props.label}</FormLabel>
        <FormControlLabel
          control={
            <MuiCheckbox checked={props.checked} onChange={props.onChange} disabled={props.disabled}></MuiCheckbox>
          }
          label={props.helperText}
        ></FormControlLabel>

        <FormHelperText>{props.errorMessage}</FormHelperText>
      </FormControl>
    </div>
  )
}
