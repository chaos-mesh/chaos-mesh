import { Box, BoxProps, Paper as MUIPaper, PaperProps as MUIPaperProps } from '@material-ui/core'

import React from 'react'

interface PaperProps extends MUIPaperProps {
  padding?: number
  boxProps?: BoxProps
}

const Paper: React.FC<PaperProps> = ({ padding = 4.5, boxProps, children, ...rest }) => (
  <MUIPaper {...rest} variant="outlined">
    <Box height="100%" p={padding} {...boxProps}>
      {children}
    </Box>
  </MUIPaper>
)

export default Paper
