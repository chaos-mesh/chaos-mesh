import React from 'react'
import { Box, Typography } from '@material-ui/core'

const PageTitle: React.FC = ({ children }) => (
  <Box ml={1.5} mb={3}>
    <Typography variant="h6">{children}</Typography>
  </Box>
)

export default PageTitle
