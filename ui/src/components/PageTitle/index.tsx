import React from 'react'
import { Box, Typography } from '@material-ui/core'

const PageTitle: React.FC = ({ children }) => (
  <Box m={3} mt={1.5}>
    <Typography variant="h6">{children}</Typography>
  </Box>
)

export default PageTitle
