import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { getIn, useFormikContext } from 'formik'

import Autocomplete from '@material-ui/lab/Autocomplete'
import React from 'react'
import T from 'components/T'

interface AutocompleteMultipleFieldProps {
  options: string[]
}

const AutocompleteMultipleField: React.FC<AutocompleteMultipleFieldProps & TextFieldProps> = ({
  options,
  ...props
}) => {
  const { values, setFieldValue } = useFormikContext()

  const name = props.name!
  const labels: string[] = getIn(values, name)
  const setLabels = (labels: string[]) => setFieldValue(name, labels)

  const onChange = (_: any, newVal: string[], reason: string) => {
    if (reason === 'clear') {
      setLabels([])

      return
    }

    setLabels(newVal)
  }

  const onDelete = (val: string) => () => setLabels(labels.filter((d) => d !== val))

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
