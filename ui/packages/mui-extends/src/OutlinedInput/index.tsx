import { OutlinedInput as MuiOutlinedInput, OutlinedInputProps } from '@mui/material'

import { forwardRef } from 'react'

const OutlinedInput = forwardRef(({ sx, ...rest }: OutlinedInputProps, ref) => (
  <MuiOutlinedInput
    size="small"
    sx={{
      px: rest.startAdornment || rest.endAdornment ? 2 : 0,
      borderColor: 'divider',
      typography: 'body2',
      '&:hover': {
        bgcolor: 'action.hover',
        '.MuiOutlinedInput-notchedOutline': {
          borderColor: 'divider',
        },
      },
      '&.Mui-focused': {
        bgcolor: 'none',
        '.MuiOutlinedInput-notchedOutline': {
          borderColor: 'primary.main',
          boxShadow: (theme) => `0 0 0 2px ${theme.palette.secondaryContainer.main}`,
        },
      },
      ...sx,
    }}
    {...rest}
    ref={ref}
  />
))

export default OutlinedInput
