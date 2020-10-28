import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { getIn, useFormikContext } from 'formik'

import Autocomplete from '@material-ui/lab/Autocomplete'
import React from 'react'
import T from 'components/T'

interface AutocompleteMultipleFieldProps {
  options: string[]
  onChangeCallback?: (labels: string[]) => void
}

const AutocompleteMultipleField: React.FC<AutocompleteMultipleFieldProps & TextFieldProps> = ({
  options,
  onChangeCallback,
  ...props
}) => {
  const { values, setFieldValue } = useFormikContext()

  const name = props.name!
  const labels: string[] = getIn(values, name)
  const setLabels = (labels: string[]) => setFieldValue(name, labels)

  const onChange = (_: any, newVal: string[], reason: string) => {
    console.log(newVal)
    if (reason === 'clear') {
      setLabels([])

      return
    }

    setLabels(newVal)
    typeof onChangeCallback === 'function' && onChangeCallback(newVal)
  }

  const onDelete = (val: string) => () => {
    const updated = labels.filter((d) => d !== val)

    setLabels(updated)
    typeof onChangeCallback === 'function' && onChangeCallback(updated)
  }

  return (
    <Box mb={3}>
      <Autocomplete
        multiple
        options={options}
        noOptionsText={T('common.noOptions')}
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
