import { Box, Typography } from '@material-ui/core'

import Paper from 'components-mui/Paper'
import React from 'react'

const NumberOf: React.FC<{ title: string | JSX.Element; num: number }> = ({ title, num }) => (
  <Paper>
    <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100px" my={6}>
      <Typography variant="overline" style={{ textAlign: 'center' }}>
        {title}
      </Typography>
      <Box mt={6}>
        <Typography variant="h5">{num}</Typography>
      </Box>
    </Box>
  </Paper>
)

export default NumberOf
