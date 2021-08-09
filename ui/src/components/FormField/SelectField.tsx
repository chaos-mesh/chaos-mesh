import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { Field, getIn, useFormikContext } from 'formik'

import { Experiment } from 'components/NewExperiment/types'

const SelectField: React.FC<TextFieldProps & { multiple?: boolean }> = ({ multiple = false, ...props }) => {
  const { values, setFieldValue } = useFormikContext<Experiment>()

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
