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

import { InputAdornment, MenuItem } from '@mui/material'
import { SelectField, TextField } from '../../FormField'
import { getIn, useFormikContext } from 'formik'

import React from 'react'
import i18n from '../../T'

const modes = [
  { name: 'Random One', value: 'one' },
  { name: 'Fixed Number', value: 'fixed' },
  { name: 'Fixed Percent', value: 'fixed-percent' },
  { name: 'Random Max Percent', value: 'random-max-percent' },
]
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

interface ModeProps {
  modeScope: string
  scope: string
  disabled: boolean
}

const Mode: React.FC<ModeProps> = ({ disabled, modeScope, scope }) => {
  const { values } = useFormikContext()
  return (
    <>
      <SelectField
        name={`${modeScope}.mode`}
        label={i18n('newE.scope.mode')}
        helperText={i18n('newE.scope.modeHelper')}
        disabled={disabled}
      >
        <MenuItem value="all">All</MenuItem>
        {modes.map((option) => (
          <MenuItem key={option.value} value={option.value}>
            {option.name}
          </MenuItem>
        ))}
      </SelectField>

      {!['all', 'one'].includes(getIn(values, modeScope).mode) && (
        <TextField
          name={`${modeScope}.value`}
          label={i18n('newE.scope.modeValue')}
          helperText={i18n('newE.scope.modeValueHelper')}
          InputProps={{
            endAdornment: modesWithAdornment.includes(getIn(values, scope).mode) && (
              <InputAdornment position="end">%</InputAdornment>
            ),
          }}
          disabled={disabled}
        />
      )}
    </>
  )
}

export default Mode
