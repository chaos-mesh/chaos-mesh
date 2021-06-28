import { Box, Table, TableBody, TableRow, Typography } from '@material-ui/core'

import ObjectConfiguration from '.'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { TableCell } from './common'

interface NodeConfigurationProps {
  template: any
}

const Suspend = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <TableRow>
        <TableCell>{T('common.name')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.name}
          </Typography>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell>{T('newE.target.kind')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.templateType}
          </Typography>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell>{T('newW.node.deadline')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.deadline}
          </Typography>
        </TableCell>
      </TableRow>
    </TableBody>
  </Table>
)

const NodeConfiguration: React.FC<NodeConfigurationProps> = ({ template: t }) => {
  const rendered = () => {
    switch (t.templateType) {
      case 'Suspend':
        return <Suspend template={t} />
      default:
        return <ObjectConfiguration config={t} inNode vertical />
    }
  }

  return (
    <Box p={3}>
      <PaperTop title={T('common.configuration')} boxProps={{ mb: 3 }} />
      {rendered()}
    </Box>
  )
}

export default NodeConfiguration
