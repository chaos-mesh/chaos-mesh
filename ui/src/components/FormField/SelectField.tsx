import React, { FC } from 'react'
import { Box, Select, SelectProps, InputLabel, FormHelperText } from '@material-ui/core'

const SelectField: FC<SelectProps & { helperText?: string }> = ({
  children,
  id,
  label,
  helperText,
  fullWidth = true,
  ...selectProps
}) => {
  return (
    <Box mb={4}>
      {label && <InputLabel id={`${id}-label`}>{label}</InputLabel>}
      <Select id={id} fullWidth={fullWidth} {...selectProps}>
        {children}
      </Select>
      {helperText && <FormHelperText id={`${id}-helper-text`}>{helperText}</FormHelperText>}
    </Box>
  )
}

export default SelectField
