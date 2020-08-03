import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { getIn, useFormikContext } from 'formik'

import Autocomplete from '@material-ui/lab/Autocomplete'
import { Experiment } from 'components/NewExperiment/types'

interface AutocompleteMultipleFieldProps {
  options: string[]
  onChangeCallback?: (labels: string[]) => void
}

const AutocompleteMultipleField: React.FC<AutocompleteMultipleFieldProps & TextFieldProps> = ({
  options,
  onChangeCallback,
  ...props
}) => {
  const { values, setFieldValue } = useFormikContext<Experiment>()

  const labelsRef = useRef(getIn(values, props.name!))
  const [labels, _setLabels] = useState<string[]>(labelsRef.current)
  const setLabels = (newVal: string[]) => {
    labelsRef.current = newVal
    _setLabels(labelsRef.current)
  }

  // For performance consider, setFieldValue before compoennt unmount
  useEffect(
    () => () => setFieldValue(props.name!, labelsRef.current),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  const onChange = (_: any, newVal: string | string[] | null, reason: string) => {
    if (reason === 'clear') {
      setLabels([])

      return
    }

    if (newVal) {
      setLabels(newVal as string[])
    }
  }

  useEffect(() => {
    if (typeof onChangeCallback === 'function') {
      onChangeCallback(labels)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [labels])

  const onDelete = (val: string) => () => setLabels(labels.filter((d) => d !== val))

  return (
    <Box mb={2}>
      <Autocomplete
        multiple
        options={options}
        value={labels}
        onChange={onChange}
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
            InputProps={{
              ...params.InputProps,
              ...props.InputProps,
              style: { paddingTop: 8 },
            }}
          />
        )}
      />
    </Box>
  )
}

export default AutocompleteMultipleField
