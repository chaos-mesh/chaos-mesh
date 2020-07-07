import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { getIn, useFormikContext } from 'formik'

import Autocomplete from '@material-ui/lab/Autocomplete'
import { Experiment } from 'components/NewExperiment/types'

interface LabelFieldProps {
  isKV?: boolean
}

const LabelField: React.FC<LabelFieldProps & TextFieldProps> = ({ isKV = false, ...props }) => {
  const { values, setFieldValue } = useFormikContext<Experiment>()

  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const labelsRef = useRef(getIn(values, props.name!))
  const [labels, setLabels] = useState<string[]>(labelsRef.current)

  useEffect(
    () => () => setFieldValue(props.name!, labelsRef.current),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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
        labelsRef.current = labels.concat([t])
        setLabels(labelsRef.current)

        if (error) {
          setError('')
        }
      }

      setText('')
    }
  }

  const onDelete = (val: string) => () => {
    labelsRef.current = labels.filter((d) => d !== val)
    setLabels(labelsRef.current)
  }

  return (
    <Box mb={2}>
      <Autocomplete
        multiple
        options={labels}
        value={labels}
        // make popup always closed
        open={false}
        forcePopupIcon={false}
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
            margin="dense"
            fullWidth
            variant="outlined"
            helperText={
              error !== ''
                ? error
                : isKV
                ? 'Type key:value and end with a space to generate a key/value pair'
                : props.helperText
            }
            onChange={onChange}
            onKeyDown={onKeyDown}
            error={error !== ''}
          />
        )}
      />
    </Box>
  )
}

export default LabelField
