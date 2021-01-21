import Paper from '../Paper'
import React from 'react'

/**
 * PaperContainer usually be used to replace the default container.
 *
 * For example:
 *
 * <TableContainer component={PaperContainer}>
 * ...
 * </TableContainer>
 *
 * @param {React.ReactNode} { children }
 */
const PaperContainer: React.FC = ({ children }) => (
  <Paper padding={false} style={{ maxHeight: 768, overflow: 'scroll' }}>
    {children}
  </Paper>
)

export default PaperContainer
