import Paper from '../Paper'

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
  <Paper sx={{ maxHeight: 768, p: 0, overflow: 'scroll' }}>{children}</Paper>
)

export default PaperContainer
