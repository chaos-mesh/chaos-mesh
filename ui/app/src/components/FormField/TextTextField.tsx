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
import Space from '@/mui-extends/Space'
import AddCircleTwoToneIcon from '@mui/icons-material/AddCircleTwoTone'
import RemoveCircleTwoToneIcon from '@mui/icons-material/RemoveCircleTwoTone'
import { Box, Button, FormHelperText, IconButton, Typography } from '@mui/material'
import { getIn, useFormikContext } from 'formik'
import _ from 'lodash'

import LabelField from './LabelField'
import TextField from './TextField'

interface TextTextFieldProps {
  name: string
  label: React.ReactNode
  helperText?: React.ReactNode
  valueLabeled?: boolean
}

export default function TextTextField({ name, label, helperText, valueLabeled }: TextTextFieldProps) {
  const { values, setFieldValue } = useFormikContext()
  const fieldValue = getIn(values, name)
  const entries = Object.entries(fieldValue)

  const handleAddKV = (n: number) => () => {
    setFieldValue(name, {
      ...fieldValue,
      ['key' + n]: { key: '', value: valueLabeled ? [] : '' },
    })
  }

  const handleRemoveKV = (key: string) => () => {
    setFieldValue(name, _.omit(fieldValue, key))
  }

  return (
    <>
      <Box>
        <Typography variant="body2" fontWeight={500}>
          {label}
        </Typography>
        {helperText && <FormHelperText>{helperText}</FormHelperText>}
      </Box>
      {entries.length > 0 ? (
        entries.map(([key], i) => (
          <Space key={key} direction="row">
            <TextField fast name={`${name}.${key}.key`} />
            {valueLabeled ? (
              <LabelField name={`${name}.${key}.value`} sx={{ width: 300 }} />
            ) : (
              <TextField fast name={`${name}.${key}.value`} />
            )}
            <IconButton color="error" onClick={handleRemoveKV(key)}>
              <RemoveCircleTwoToneIcon />
            </IconButton>
            {i === entries.length - 1 && (
              <IconButton color="primary" onClick={handleAddKV(i + 1)}>
                <AddCircleTwoToneIcon />
              </IconButton>
            )}
          </Space>
        ))
      ) : (
        <Box>
          <Button variant="contained" onClick={handleAddKV(0)}>
            {`Add Key/Value${valueLabeled ? 's' : ''} Pair`}
          </Button>
        </Box>
      )}
    </>
  )
}
