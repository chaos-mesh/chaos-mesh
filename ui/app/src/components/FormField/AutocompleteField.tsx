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
import { getIn, useFormikContext } from 'formik'

import MuiExtendsAutocompleteField from '@ui/mui-extends/esm/AutocompleteField'
import type { AutocompleteFieldProps as MuiExtendsAutocompleteFieldProps } from '@ui/mui-extends/esm/AutocompleteField'

import { T } from 'components/T'

export interface AutocompleteFieldProps extends MuiExtendsAutocompleteFieldProps {
  name: string
}

const AutocompleteField: React.FC<AutocompleteFieldProps> = ({ name, multiple, options, ...props }) => {
  const { values, setFieldValue } = useFormikContext()
  const value = getIn(values, name) || []

  const onChange = (_: any, newVal: string | string[] | null, reason: string) => {
    if (reason === 'clear') {
      setFieldValue(name, multiple ? [] : '')

      return
    }

    setFieldValue(name, newVal)
  }

  const onDelete = (val: string) => () =>
    setFieldValue(
      name,
      value.filter((d: string) => d !== val)
    )

  return (
    <MuiExtendsAutocompleteField
      name={name}
      {...props}
      multiple={multiple}
      options={!props.disabled ? options : []}
      noOptionsText={<T id="common.noOptions" />}
      value={value}
      onChange={onChange}
      onRenderValueDelete={onDelete}
    />
  )
}

export default AutocompleteField
