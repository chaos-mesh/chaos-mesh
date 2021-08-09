import {
  ListItemProps,
  ListProps,
  List as MUIList,
  ListItem as MUIListItem,
  TableCell as MUITableCell,
  Table,
  TableBody,
  TableCellProps,
  TableRow,
  Typography,
} from '@material-ui/core'

import { ExperimentKind } from 'components/NewExperiment/types'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { objToArrBySep } from 'lib/utils'

export const TableCell = (props: TableCellProps) => (
  <MUITableCell sx={{ borderBottom: 'none', '&:first-child': { width: '50%' } }} {...props} />
)
export const List = (props: ListProps) => <MUIList sx={{ p: 0 }} {...props} />
export const ListItem = (props: ListItemProps) => <MUIListItem sx={{ p: 0 }} {...props} />

export const Selector = ({ data }: any) => (
  <Table size="small">
    <TableBody>
      {data.namespaces && (
        <TableRow>
          <TableCell>{T('k8s.namespaceSelectors')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {data.namespaces.join('; ')}
            </Typography>
          </TableCell>
        </TableRow>
      )}
      {data.labelSelectors && (
        <TableRow>
          <TableCell>{T('k8s.labelSelectors')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {objToArrBySep(data.labelSelectors, ': ').join('; ')}
            </Typography>
          </TableCell>
        </TableRow>
      )}
      {data.annotationSelectors && (
        <TableRow>
          <TableCell>{T('k8s.annotationSelectors')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {objToArrBySep(data.annotationSelectors, ': ').join('; ')}
            </Typography>
          </TableCell>
        </TableRow>
      )}
    </TableBody>
  </Table>
)

const Pod = ({ data: { containerName } }: any) => (
  <>
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

export const Stress = ({ data: { stressors, stressngStressors, containerName } }: any) => (
  <>
    {stressors.cpu && (
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
      </>
    )}
    {stressors.memory && (
      <>
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
      </>
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

export const Experiment = ({ kind, data }: { kind: ExperimentKind; data: any }) => (
  <Table size="small">
    <TableBody>
      <TableRow>
        <TableCell>{T('newE.target.kind')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {kind}
          </Typography>
        </TableCell>
      </TableRow>
      {['PodChaos', 'NetworkChaos', 'IOChaos'].includes(kind) && (
        <TableRow>
          <TableCell>{T('newE.target.action')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {data.action}
            </Typography>
          </TableCell>
        </TableRow>
      )}
      {kind === 'PodChaos' && <Pod data={data} />}
      {kind === 'NetworkChaos' && <Network data={data} />}
      {kind === 'IOChaos' && <IO data={data} />}
      {kind === 'StressChaos' && <Stress data={data} />}
      {kind === 'TimeChaos' && <Time data={data} />}
    </TableBody>
  </Table>
)
