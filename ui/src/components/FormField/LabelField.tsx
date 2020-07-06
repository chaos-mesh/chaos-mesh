import { Box, Chip, TextFieldProps } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Experiment } from 'components/NewExperiment/types'
import TextField from './TextField'
import { useFormikContext } from 'formik'

interface LabelFieldProps {
  isKV?: boolean
}

const LabelField: React.FC<LabelFieldProps & TextFieldProps> = ({ children, isKV = false, ...props }) => {
  const { setFieldValue } = useFormikContext<Experiment>()

  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const [labels, setLabels] = useState<string[]>([])

  useEffect(() => {
    setFieldValue(props.name!, labels)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [labels])

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => setText(e.target.value)

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === ' ') {
      const t = text.trim()

      if (isKV && !/^[\w-]+:[\w-]+$/.test(t)) {
        setError('Invalid key:value format')

        return
      }

      const duplicate = labels.some((d) => d === t)

      setText('')

      if (!duplicate) {
        setLabels(labels.concat([t]))

        if (error) {
          setError('')
        }
      }
    }
  }

  const onDelete = (val: string) => () => setLabels(labels.filter((d) => d !== val))

  return (
    <Box mb={2}>
      <TextField
        {...props}
        helperText={
          error !== ''
            ? error
            : isKV
            ? 'Type key:value and end with a space to generate a key/value pair'
            : props.helperText
        }
        value={text}
        onChange={onChange}
        onKeyDown={onKeyDown}
        error={error !== ''}
      >
        {children}
      </TextField>
      <Box display="flex" flexWrap="wrap">
        {labels.map((val) => (
          <Box key={val} m={0.5}>
            <Chip label={val} color="primary" style={{ height: 24 }} clickable onDelete={onDelete(val)} />
          </Box>
        ))}
      </Box>
    </Box>
  )
}

export default LabelField
