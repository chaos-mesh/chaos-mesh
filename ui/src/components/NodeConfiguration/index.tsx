import {
  Box,
  List,
  ListItem as MUIListItem,
  TableCell as MUITableCell,
  Table,
  TableBody,
  TableRow,
  Typography,
} from '@material-ui/core'
import { objToArrBySep, toCamelCase, toTitleCase } from 'lib/utils'

import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { withStyles } from '@material-ui/styles'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

const ListItem: any = withStyles({
  root: {
    paddingLeft: 0,
  },
})(MUIListItem)

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
          {t.templateType}
        </Typography>
      </TableCell>
    </TableRow>
    <TableRow>
      <TableCell>{T('newE.run.duration')}</TableCell>
      <TableCell>
        <Typography variant="body2" color="textSecondary">
          {t.duration}
        </Typography>
      </TableCell>
    </TableRow>
  </>
)

const Selector = ({ data }: any) => (
  <>
    <TableRow>
      <TableCell>{T('k8s.labelSelectors')}</TableCell>
      <TableCell>
        <List>
          {objToArrBySep(data.selector.labelSelectors, ': ').map((d) => (
            <ListItem key={d}>
              <Typography variant="body2" color="textSecondary">
                {d}
              </Typography>
            </ListItem>
          ))}
        </List>
      </TableCell>
    </TableRow>
  </>
)

const Pod = ({ data: { action, containerName } }: any) => (
  <>
    <TableRow>
      <TableCell>{T('newE.target.action')}</TableCell>
      <TableCell>
        <Typography variant="body2" color="textSecondary">
          {action.includes('-')
            ? (function () {
                const split = action.split('-')

                return toTitleCase(split[0]) + ' ' + toTitleCase(split[1])
              })()
            : toTitleCase(action)}
        </Typography>
      </TableCell>
    </TableRow>
    {containerName && (
      <TableRow>
        <TableCell>Container name</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {containerName}
          </Typography>
        </TableCell>
      </TableRow>
    )}
  </>
)

const Network = ({ data }: any) => (
  <>
    <TableRow>
      <TableCell>{T('newE.target.action')}</TableCell>
      <TableCell>
        <Typography variant="body2" color="textSecondary">
          {data.action}
        </Typography>
      </TableCell>
    </TableRow>
    {data.action && (
      <TableRow>
        <TableCell>{toTitleCase(data.action)}</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(data[data.action], ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    {data.direction && (
      <TableRow>
        <TableCell>Direction</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.direction}
          </Typography>
        </TableCell>
      </TableRow>
    )}
  </>
)

const IO = ({ data }: any) => (
  <>
    <TableRow>
      <TableCell>{T('newE.target.action')}</TableCell>
      <TableCell>
        <Typography variant="body2" color="textSecondary">
          {data.action}
        </Typography>
      </TableCell>
    </TableRow>
    {data.delay && (
      <TableRow>
        <TableCell>Delay</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.delay}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.errno && (
      <TableRow>
        <TableCell>Errno</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.errno}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.attr && (
      <TableRow>
        <TableCell>Attr</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(data.attr, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    {data.volumePath && (
      <TableRow>
        <TableCell>Volume path</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.volumePath}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.path && (
      <TableRow>
        <TableCell>Path</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.path}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.containerName && (
      <TableRow>
        <TableCell>Container name</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.containerName}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.percent && (
      <TableRow>
        <TableCell>Percent</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.percent}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.methods && (
      <TableRow>
        <TableCell>Methods</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(data.methods, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
  </>
)

const Stress = ({ data: { stressors, stressngStressors, containerName } }: any) => (
  <>
    <PaperTop title="CPU" />
    {stressors.cpu.workers && (
      <TableRow>
        <TableCell>Workers</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {stressors.cpu.workers}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressors.cpu.load && (
      <TableRow>
        <TableCell>Load</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {stressors.cpu.load}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressors.cpu.options && (
      <TableRow>
        <TableCell>Options</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(stressors.cpu.options, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    <PaperTop title="Memory" />
    {stressors.memory.workers && (
      <TableRow>
        <TableCell>Workers</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {stressors.memory.workers}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressors.memory.size && (
      <TableRow>
        <TableCell>Size</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {stressors.memory.size}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressors.memory.options && (
      <TableRow>
        <TableCell>Options</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(stressors.memory.options, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    {stressors.stressngStressors && (
      <TableRow>
        <TableCell>Options of stress-ng</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(stressors.stressngStressors, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    {containerName && (
      <TableRow>
        <TableCell>Container name</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {containerName}
          </Typography>
        </TableCell>
      </TableRow>
    )}
  </>
)

const Time = ({ data }: any) => (
  <>
    {data.timeOffset && (
      <TableRow>
        <TableCell>Time offset</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {data.timeOffset}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {data.clockIds && (
      <TableRow>
        <TableCell>Clock ids</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(data.clockIds, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
    {data.containerNames && (
      <TableRow>
        <TableCell>Container names</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(data.containerNames, ': ').map((d) => (
              <ListItem key={d}>
                <Typography variant="body2" color="textSecondary">
                  {d}
                </Typography>
              </ListItem>
            ))}
          </List>
        </TableCell>
      </TableRow>
    )}
  </>
)

const Suspend = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <Common template={t} />
    </TableBody>
  </Table>
)

const Experiment = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <Common template={t} />
      <Selector data={t[toCamelCase(t.templateType)]} />
      {t.podChaos && <Pod data={t.podChaos} />}
      {t.networkChaos && <Network data={t.networkChaos} />}
      {t.ioChaos && <IO data={t.ioChaos} />}
      {t.stressChaos && <Stress data={t.stressChaos} />}
      {t.timeChaos && <Time data={t.timeChaos} />}
    </TableBody>
  </Table>
)

const NodeConfiguration: React.FC<NodeConfigurationProps> = ({ template: t }) => {
  const rendered = () => {
    switch (t.templateType) {
      case 'Suspend':
        return <Suspend template={t} />
      default:
        return <Experiment template={t} />
    }
  }

  return (
    <Box p={3}>
      <PaperTop title={T('common.configuration')} />
      {rendered()}
    </Box>
  )
}

export default NodeConfiguration
