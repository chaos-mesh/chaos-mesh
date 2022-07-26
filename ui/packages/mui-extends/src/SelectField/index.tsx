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
import { Box, Chip, Select } from '@mui/material'
import type { SelectProps, TextFieldProps } from '@mui/material'

import FormControl from '../FormControl'
import OutlinedInput from '../OutlinedInput'

export type SelectFieldProps<T = string> = SelectProps<T> & {
  label?: TextFieldProps['label']
  helperText?: TextFieldProps['helperText']
  onRenderValueDelete?: (value: string) => (event: any) => void
}

export default function SelectField<T>({ label, helperText, onRenderValueDelete, ...props }: SelectFieldProps<T>) {
  const { disabled, error, fullWidth } = props

  return (
    <FormControl disabled={disabled} error={error} label={label} helperText={helperText} fullWidth={fullWidth}>
      <Select
        {...props}
        input={<OutlinedInput />}
        renderValue={
          props.multiple
            ? (selected: unknown) => (
                <Box display="flex" flexWrap="wrap" gap={1}>
                  {(selected as string[]).map((val) => (
                    <Chip
                      key={val}
                      label={val}
                      color="primary"
                      onDelete={onRenderValueDelete ? onRenderValueDelete(val) : undefined}
                      onMouseDown={(e) => e.stopPropagation()}
                      sx={{ height: 24 }}
                    />
                  ))}
                </Box>
              )
            : undefined
        }
      />
    </FormControl>
  )
}
