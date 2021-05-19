import { Box, Paper as MUIPaper, PaperProps as MUIPaperProps } from '@material-ui/core'

import React from 'react'

interface PaperProps extends MUIPaperProps {
  padding?: number
}

const Paper: React.FC<PaperProps> = ({ padding = 4.5, children, ...rest }) => (
  <MUIPaper {...rest} variant="outlined">
    <Box height="100%" p={padding}>
      {children}
    </Box>
  </MUIPaper>
)

export default Paper
