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

import { Box, Chip, TextFieldProps } from '@mui/material'

import TextField from '../TextField'

export type SelectFieldProps = TextFieldProps & {
  multiple?: boolean
  onChipDelete?: (value: string) => (event: any) => void
}

export default ({ multiple = false, onChipDelete, ...props }: SelectFieldProps) => {
  const selectProps = {
    ...props.SelectProps,
    multiple,
    renderValue: multiple
      ? (selected: unknown) => (
          <Box display="flex" flexWrap="wrap" mt={1}>
            {(selected as string[]).map((val) => (
              <Chip
                key={val}
                label={val}
                color="primary"
                onDelete={onChipDelete ? onChipDelete(val) : undefined}
                onMouseDown={(e) => e.stopPropagation()}
                style={{ height: 24, margin: 1 }}
              />
            ))}
          </Box>
        )
      : undefined,
  }

  return <TextField {...props} select SelectProps={selectProps} />
}
