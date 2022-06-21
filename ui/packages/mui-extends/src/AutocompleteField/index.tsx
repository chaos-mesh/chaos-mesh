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
import { Autocomplete, Chip, Paper } from '@mui/material'
import type { AutocompleteProps, TextFieldProps } from '@mui/material'

import FormControl from '../FormControl'
import OutlinedInput from '../OutlinedInput'

export interface AutocompleteFieldProps<
  T = string,
  Multiple extends boolean | undefined = boolean,
  DisableClearable extends boolean | undefined = boolean,
  FreeSolo extends boolean | undefined = boolean
> extends Omit<AutocompleteProps<T, Multiple, DisableClearable, FreeSolo>, 'renderInput'> {
  name?: string
  label?: TextFieldProps['label']
  helperText?: TextFieldProps['helperText']
  error?: boolean
  onRenderValueDelete?: (value: T) => (event: any) => void
}

export default function AutocompleteField<T>({
  name,
  label,
  helperText,
  error,
  onRenderValueDelete,
  ...props
}: AutocompleteFieldProps<T>) {
  const { disabled, fullWidth } = props

  return (
    <FormControl
      disabled={disabled}
      error={error}
      label={label}
      LabelProps={{ htmlFor: name }}
      helperText={helperText}
      fullWidth={fullWidth}
    >
      <Autocomplete
        id={name}
        {...props}
        renderInput={(params) => (
          <OutlinedInput
            name={name}
            {...params.InputProps}
            inputProps={params.inputProps}
            error={error}
            sx={{ width: '100%' }}
          />
        )}
        renderTags={
          props.multiple
            ? (value: T[], getTagProps) =>
                value.map((val: T, index: number) => {
                  const tagProps = getTagProps({ index })

                  return (
                    <Chip
                      {...tagProps}
                      label={val}
                      color="primary"
                      onDelete={onRenderValueDelete ? onRenderValueDelete(val) : tagProps.onDelete}
                      sx={{ height: 24 }}
                    />
                  )
                })
            : undefined
        }
        PaperComponent={(props) => <Paper {...props} sx={{ mt: 1 }} />}
      />
    </FormControl>
  )
}
