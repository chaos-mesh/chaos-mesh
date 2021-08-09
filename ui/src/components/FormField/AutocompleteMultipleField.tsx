import { Autocomplete, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { getIn, useFormikContext } from 'formik'

import Paper from 'components-mui/Paper'
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
    <Autocomplete
      multiple
      options={!props.disabled ? options : []}
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
      renderInput={(params) => <TextField {...params} {...props} size="small" fullWidth />}
      PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
    />
  )
}

export default AutocompleteMultipleField
