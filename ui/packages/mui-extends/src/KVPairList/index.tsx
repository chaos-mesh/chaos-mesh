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

import React from 'react'

export type KVPair<K, V> = {
  key: K
  value?: V
}

export type KVPairListProps<K, V> = {
  label: string
  helperText: string
  error: boolean
  data: Array<KVPair<K, V>>
  required: boolean
}
export default <K, V>({ label, helperText, error, required, data }: KVPairListProps<K, V>) => {
  return (
    <FormControl error={error} required={required}>
      <FormLabel>{label}</FormLabel>
      <div>{helperText}</div>
      {data.map((item) => {
        return <KVPairRow data={item}></KVPairRow>
      })}
    </FormControl>
  )
}

export type KVPairRowProps<K extends unknown, V extends unknown> = {
  data: KVPair<K, V>
  renderKey?: (key: K) => JSX.Element
  renderValue?: (value?: V) => JSX.Element
}

export const KVPairRow = <K, V>({ data, renderKey, renderValue }: KVPairRowProps<K, V>) => {
  return (
    <>
      {renderKey != null ? renderKey(data.key) : <div>{data.key}</div>}
      {renderValue != null ? renderValue(data.value) : <div>{data.value}</div>}
    </>
  )
}
