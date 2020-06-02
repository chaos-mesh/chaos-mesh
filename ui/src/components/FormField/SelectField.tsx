import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import React, { FC } from 'react'

const SelectField: FC<TextFieldProps & { multiple?: boolean }> = ({
  children,
  fullWidth = true,
  multiple = false,
  ...selectProps
}) => {
  const SelectProps = {
    multiple,
    renderValue: multiple
      ? (selected: any) => (
          <Box display="flex" flexWrap="wrap">
            {(selected as string[]).map((value) => (
              <Box key={value} m={0.5}>
                <Chip label={value} color="primary" />
              </Box>
            ))}
          </Box>
        )
      : undefined,
  }

  return (
    <Box mb={2}>
      <TextField
        select
        margin="dense"
        fullWidth={fullWidth}
        variant="outlined"
        SelectProps={SelectProps}
        {...selectProps}
      >
        {children}
      </TextField>
    </Box>
  )
}

export default SelectField
