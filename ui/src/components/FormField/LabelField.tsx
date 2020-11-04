import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { getIn, useFormikContext } from 'formik'

import Autocomplete from '@material-ui/lab/Autocomplete'
import { Experiment } from 'components/NewExperiment/types'
import T from 'components/T'

interface LabelFieldProps {
  isKV?: boolean // whether to use the key: value format
}

const LabelField: React.FC<LabelFieldProps & TextFieldProps> = ({ isKV = false, ...props }) => {
  const { values, setFieldValue } = useFormikContext<Experiment>()

  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const labelsInForm = getIn(values, props.name!)
  const labelsRef = useRef(labelsInForm)
  const [labels, _setLabels] = useState<string[]>(labelsRef.current)
  const setLabels = (newVal: string[]) => {
    labelsRef.current = newVal
    _setLabels(labelsRef.current)
  }

  useEffect(
    () => () => setFieldValue(props.name!, labelsRef.current),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  useEffect(() => setLabels(labelsInForm), [labelsInForm])

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

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === ' ') {
      const t = text.trim()

      if (isKV && !/^[\w-]+:[\w-]+$/.test(t)) {
        setError('Invalid key:value format')

        return
      }

      const duplicate = labels.some((d) => d === t)

      if (!duplicate) {
        setLabels(labels.concat([t]))

        if (error) {
          setError('')
        }
      }

      setText('')
    }

    if (e.key === 'Backspace' && text === '') {
      setLabels(labels.slice(0, labels.length - 1))
    }
  }

  const onDelete = (val: string) => () => setLabels(labels.filter((d) => d !== val))

  return (
    <Box mb={2}>
      <Autocomplete
        multiple
        options={labels}
        value={labels}
        // make popup always closed
        open={false}
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
            variant="outlined"
            margin="dense"
            fullWidth
            helperText={error !== '' ? error : isKV ? T('common.isKVHelperText') : props.helperText}
            onChange={onInputChange}
            onKeyDown={onKeyDown}
            error={error !== ''}
          />
        )}
      />
    </Box>
  )
}

export default LabelField
