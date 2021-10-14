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
import { Autocomplete, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { getIn, useFormikContext } from 'formik'
import { useEffect, useState } from 'react'

import T from 'components/T'

interface LabelFieldProps {
  isKV?: boolean // whether to use the key:value format,
  errorText?: string
}

const LabelField: React.FC<LabelFieldProps & TextFieldProps> = ({ isKV = false, errorText, ...props }) => {
  const { values, setFieldValue } = useFormikContext()

  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const name = props.name!
  const labels: string[] = getIn(values, name)
  const setLabels = (labels: string[]) => setFieldValue(name, labels)

  useEffect(() => {
    if (errorText) {
      setError(errorText)
    }
  }, [errorText])

  const onChange = (_: any, __: any, reason: string) => {
    if (reason === 'clear') {
      setLabels([])
    }
  }

  const onInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value

    if (val === ' ') {
      setText('')

      return
    }

    setText(e.target.value)
  }

  const processText = () => {
    const t = text.trim()

    if (t === '') {
      return
    }

    if (isKV && !/^.+:.+$/.test(t)) {
      setError('Invalid key:value format')

      return
    }

    const duplicate = labels.some((d) => d === t)

    if (!duplicate) {
      setLabels(labels.concat([t]))
    }

    setText('')
  }

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (error) {
      setError('')
    }

    if (e.key === ' ') {
      processText()
    }

    if (e.key === 'Backspace' && text === '') {
      setLabels(labels.slice(0, labels.length - 1))
    }
  }

  const onDelete = (val: string) => () => setLabels(labels.filter((d: string) => d !== val))

  return (
    <Autocomplete
      multiple
      options={labels}
      value={labels}
      open={false} // make popup always closed
      forcePopupIcon={false}
      onChange={onChange}
      inputValue={text}
      renderTags={(value: string[], getTagProps) =>
        value.map((val: string, index: number) => (
          <Chip
            {...getTagProps({ index })}
            style={{ height: 24 }}
            label={val}
            color="primary"
            onDelete={onDelete(val)}
          />
        ))
      }
      renderInput={(params) => (
        <TextField
          {...params}
          {...props}
          size="small"
          fullWidth
          helperText={error !== '' ? error : isKV ? T('common.isKVHelperText') : props.helperText}
          error={error !== ''}
          onChange={onInputChange}
          onKeyDown={onKeyDown}
          onBlur={processText}
        />
      )}
    />
  )
}

export default LabelField
