import { Box } from '@material-ui/core'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import React from 'react'
import T from 'components/T'
import { toTitleCase } from 'lib/utils'

interface WrapperProps {
  from?: 'yaml'
}

const Wrapper: React.FC<WrapperProps> = ({ from = '', children }) => (
  <Paper>
    <PaperTop title={T(`newE.loadFrom${toTitleCase(from)}`)} />
    <Box p={6} maxHeight={450} style={{ overflowY: 'scroll' }}>
      {children}
    </Box>
  </Paper>
)

export default Wrapper
