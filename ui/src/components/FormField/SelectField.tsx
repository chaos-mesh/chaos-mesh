import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { Field, getIn, useFormikContext } from 'formik'

import { Experiment } from 'components/NewExperiment/types'
import React from 'react'

const SelectField: React.FC<TextFieldProps & { multiple?: boolean }> = ({
  children,
  fullWidth = true,
  multiple = false,
  ...props
}) => {
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
          <Box display="flex" flexWrap="wrap">
            {(selected as string[]).map((val) => (
              <Box key={val} m={0.5}>
                <Chip
                  label={val}
                  color="primary"
                  style={{ height: 24 }}
                  clickable
                  onDelete={onDelete(val)}
                  onMouseDown={(e) => e.stopPropagation()}
                />
              </Box>
            ))}
          </Box>
        )
      : undefined,
  }

  return (
    <Box mb={2}>
      <Field
        as={TextField}
        select
        margin="dense"
        fullWidth={fullWidth}
        variant="outlined"
        SelectProps={SelectProps}
        {...props}
      >
        {children}
      </Field>
    </Box>
  )
}

export default SelectField
