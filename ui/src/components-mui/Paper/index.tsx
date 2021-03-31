import { Box, Paper as MUIPaper, PaperProps as MUIPaperProps } from '@material-ui/core'

import React from 'react'

interface PaperProps extends MUIPaperProps {
  padding?: number
}

const Paper: React.FC<PaperProps> = ({ padding = true, children, ...rest }) => (
  <MUIPaper {...rest} variant="outlined">
    <Box p={padding || 3}>{children}</Box>
  </MUIPaper>
)

export default Paper
