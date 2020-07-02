import { Box, Typography } from '@material-ui/core'

import React from 'react'

interface PaperTopProps {
  title: string
}

const PaperTop: React.FC<PaperTopProps> = ({ title, children }) => (
  <Box display="flex" justifyContent="space-between" alignItems="center" width="100%" height="64px">
    <Box ml={2}>
      <Typography variant="h6">{title}</Typography>
    </Box>
    {children}
  </Box>
)

export default PaperTop
