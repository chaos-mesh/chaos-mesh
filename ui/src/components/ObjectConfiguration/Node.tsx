import { Box, Table, TableBody, TableRow, Typography } from '@material-ui/core'

import { Branch } from 'slices/workflows'
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

const Custom = ({ template: t }: NodeConfigurationProps) => {
  const { container } = t.task

  return (
    <>
      <Typography variant="subtitle2" gutterBottom>
        {T('newE.steps.basic')}
      </Typography>
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
        </TableBody>
      </Table>
      <Typography variant="subtitle2" gutterBottom>
        {T('newW.node.container.title')}
      </Typography>
      <Table size="small">
        <TableRow>
          <TableCell>{T('common.name')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {container.name}
            </Typography>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>{T('newW.node.container.image')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {container.image}
            </Typography>
          </TableCell>
        </TableRow>
        {container.command && (
          <TableRow>
            <TableCell>{T('newW.node.container.command')}</TableCell>
            <TableCell>
              {container.command.map((d: string, i: number) => (
                <Typography key={i} variant="body2" color="textSecondary">
                  - {d}
                </Typography>
              ))}
            </TableCell>
          </TableRow>
        )}
      </Table>
      <Typography variant="subtitle2" gutterBottom>
        {T('newW.node.conditionalBranches.title')}
      </Typography>
      {t.conditionalBranches.map((d: Branch, i: number) => (
        <Box key={i}>
          <Typography variant="subtitle2" gutterBottom>
            {T('newW.node.conditionalBranches.branch')} {i + 1}
          </Typography>
          <Table size="small">
            <TableRow>
              <TableCell>{T('newW.node.conditionalBranches.target')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {d.target}
                </Typography>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>{T('newW.node.conditionalBranches.expression')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {d.expression}
                </Typography>
              </TableCell>
            </TableRow>
          </Table>
        </Box>
      ))}
    </>
  )
}

const NodeConfiguration: React.FC<NodeConfigurationProps> = ({ template: t }) => {
  const rendered = () => {
    switch (t.templateType) {
      case 'Suspend':
        return <Suspend template={t} />
      case 'Task':
        return <Custom template={t} />
      default:
        return <ObjectConfiguration config={t} inNode vertical />
    }
  }

  return (
    <Box p={4.5}>
      <PaperTop title={T('common.configuration')} boxProps={{ mb: 3 }} />
      {rendered()}
    </Box>
  )
}

export default NodeConfiguration
