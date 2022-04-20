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

import { Box, FormControl, FormControlLabel, FormHelperText, FormLabel } from '@mui/material'

import React from 'react'
import TextField from '../TextField/index'

export type KVPairListProps = {
  name: string
  value: any
  label: string
  onChange: (event: any) => {}
}
export default ({ name, value, label, onChange }: KVPairListProps) => {
  const data = value

  console.log('rendering kv list pair')
  console.log('data')
  console.log(data)

  function fieldChange(k: string, v: any) {
    onChange(v)
  }

  return (
    <FormControl>
      <FormLabel>{label}</FormLabel>

      {((value) => {
        const valueCopy = { ...value }
        var items: JSX.Element[] = []
        var i = 0
        for (const k in valueCopy) {
          const v = valueCopy[k]
          const item = (
            <Box key={i} sx={{ display: 'flex' }}>
              <TextField
                onChange={(event) => {
                  var newK = event.target.value
                  var newData: any = {}
                  for (const iterK in data) {
                    if (iterK === k) {
                      newData[newK] = data[k]
                    } else {
                      newData[iterK] = data[iterK]
                    }
                  }
                  event.target.name = newK
                  fieldChange(newData, event)
                  console.log(data)
                  console.log(newData)
                }}
                defaultValue={k}
              ></TextField>
              <TextField
                onChange={(event) => {
                  fieldChange(k, event)
                }}
                name={`${name}.${k}`}
                defaultValue={v}
              ></TextField>
            </Box>
          )
          items.push(item)
          i += 1
        }
        return <Box>{items}</Box>
      })(data)}
    </FormControl>
  )
}
