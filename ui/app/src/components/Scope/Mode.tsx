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
import { InputAdornment, MenuItem } from '@mui/material'
import { getIn, useFormikContext } from 'formik'

import { SelectField, TextField } from '@/components/FormField'
import { T } from '@/components/T'

const modes = [
  { name: 'Random One', value: 'one' },
  { name: 'Fixed Number', value: 'fixed' },
  { name: 'Fixed Percent', value: 'fixed-percent' },
  { name: 'Random Max Percent', value: 'random-max-percent' },
]
const modesWithAdornment = ['fixed-percent', 'random-max-percent']

interface ModeProps {
  modeScope: string
}

const Mode: ReactFCWithChildren<ModeProps> = ({ modeScope }) => {
  const { values } = useFormikContext()
  const modePath = modeScope ? `${modeScope}.mode` : 'mode'
  const valuePath = modeScope ? `${modeScope}.value` : 'value'
  const mode = getIn(values, modePath)

  return (
    <>
      <SelectField name={modePath} label={<T id="newE.scope.mode" />} helperText={<T id="newE.scope.modeHelper" />}>
        <MenuItem value="all">All</MenuItem>
        {modes.map((option) => (
          <MenuItem key={option.value} value={option.value}>
            {option.name}
          </MenuItem>
        ))}
      </SelectField>

      {!['all', 'one'].includes(mode) && (
        <TextField
          name={valuePath}
          label={<T id="newE.scope.modeValue" />}
          helperText={<T id="newE.scope.modeValueHelper" />}
          endAdornment={modesWithAdornment.includes(mode) && <InputAdornment position="end">%</InputAdornment>}
        />
      )}
    </>
  )
}

export default Mode
