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
import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { Field, getIn, useFormikContext } from 'formik'

const SelectField: React.FC<TextFieldProps & { multiple?: boolean }> = ({ multiple = false, ...props }) => {
  const { values, setFieldValue } = useFormikContext()

  const onDelete = (val: string) => () =>
    setFieldValue(
      props.name!,
      getIn(values, props.name!).filter((d: string) => d !== val)
    )

  const SelectProps = {
    multiple,
    renderValue: multiple
      ? (selected: any) => (
          <Box display="flex" flexWrap="wrap" mt={1}>
            {(selected as string[]).map((val) => (
              <Chip
                key={val}
                style={{ height: 24, margin: 1 }}
                label={val}
                color="primary"
                onDelete={onDelete(val)}
                onMouseDown={(e) => e.stopPropagation()}
              />
            ))}
          </Box>
        )
      : undefined,
  }

  const rendered = (
    <Field
      {...props}
      className={props.className}
      as={TextField}
      select
      size="small"
      fullWidth
      SelectProps={SelectProps}
    />
  )

  return rendered
}

export default SelectField
