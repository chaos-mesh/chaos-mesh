/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
} from '@mui/material'

import { ExperimentKind } from 'components/NewExperiment/types'
import i18n from 'components/T'
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
          <TableCell>{i18n('k8s.namespaceSelectors')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {data.namespaces.join('; ')}
            </Typography>
          </TableCell>
        </TableRow>
      )}
      {data.labelSelectors && (
        <TableRow>
          <TableCell>{i18n('k8s.labelSelectors')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {objToArrBySep(data.labelSelectors, ': ').join('; ')}
            </Typography>
          </TableCell>
        </TableRow>
      )}
      {data.annotationSelectors && (
        <TableRow>
          <TableCell>{i18n('k8s.annotationSelectors')}</TableCell>
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
    {stressors.cpu && stressors.cpu.workers > 0 && (
      <TableRow>
        <TableCell>CPU</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            workers: {stressors.cpu.workers}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            size: {stressors.cpu.size}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressors.memory && stressors.memory.workers > 0 && (
      <TableRow>
        <TableCell>Memory</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            workers: {stressors.memory.workers}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            size: {stressors.memory.size}
          </Typography>
        </TableCell>
      </TableRow>
    )}
    {stressngStressors && (
      <TableRow>
        <TableCell>Options of stress-ng</TableCell>
        <TableCell>
          <List>
            {objToArrBySep(stressngStressors, ': ').map((d) => (
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
        <TableCell>{i18n('newE.target.kind')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {kind}
          </Typography>
        </TableCell>
      </TableRow>
      {['PodChaos', 'NetworkChaos', 'IOChaos'].includes(kind) && (
        <TableRow>
          <TableCell>{i18n('newE.target.action')}</TableCell>
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
