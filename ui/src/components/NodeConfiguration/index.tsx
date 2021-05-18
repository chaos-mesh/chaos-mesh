import { Box, TableCell as MUITableCell, Table, TableBody, TableRow, Typography } from '@material-ui/core'

import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { withStyles } from '@material-ui/core/styles'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface NodeConfigurationProps {
  template: any
}

const Common = ({ template: t }: NodeConfigurationProps) => (
  <>
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
          {t.template_type}
        </Typography>
      </TableCell>
    </TableRow>
  </>
)

const Suspend = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <Common template={t} />
      <TableRow>
        <TableCell>{T('newE.schedule.duration')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.duration}
          </Typography>
        </TableCell>
      </TableRow>
    </TableBody>
  </Table>
)

const Experiment = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <Common template={t} />
      <TableRow>
        <TableCell>{T('newE.schedule.duration')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.duration}
          </Typography>
        </TableCell>
      </TableRow>
    </TableBody>
  </Table>
)

const NodeConfiguration: React.FC<NodeConfigurationProps> = ({ template: t }) => {
  const rendered = () => {
    switch (t.template_type) {
      case 'Suspend':
        return <Suspend template={t} />
      default:
        return <Experiment template={t} />
    }
  }

  return (
    <Box p={3}>
      <PaperTop title={T('common.configuration')}></PaperTop>
      {rendered()}
    </Box>
  )
}

export default NodeConfiguration
