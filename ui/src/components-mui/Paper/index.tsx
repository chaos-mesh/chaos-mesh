import { Box, Paper as MUIPaper, PaperProps as MUIPaperProps } from '@material-ui/core'

import React from 'react'

interface PaperProps extends MUIPaperProps {
  padding?: boolean
}

const Paper: React.FC<PaperProps> = ({ padding = true, children, ...rest }) => (
  <MUIPaper {...rest} variant="outlined">
    {padding ? <Box p={3}>{children}</Box> : children}
  </MUIPaper>
)

export default Paper
