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
import { Field, getIn, useFormikContext } from 'formik'

import MuiExtendsSelectField, { SelectFieldProps } from '@ui/mui-extends/esm/SelectField'

function SelectField<T>(props: SelectFieldProps<T>) {
  const { values, setFieldValue } = useFormikContext()

  const onDelete = (val: string) => () =>
    setFieldValue(
      props.name!,
      (getIn(values, props.name!) as string[]).filter((d) => d !== val)
    )

  return <Field {...props} as={MuiExtendsSelectField} onRenderValueDelete={onDelete} />
}

export default SelectField
