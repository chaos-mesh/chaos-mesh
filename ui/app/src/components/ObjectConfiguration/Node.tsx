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
import { Box, Table, TableBody, TableRow, Typography } from '@mui/material'

import { Branch } from 'slices/workflows'
import ObjectConfiguration from '.'
import { TableCell } from './common'
import i18n from 'components/T'

interface NodeConfigurationProps {
  template: any
}

const Suspend = ({ template: t }: NodeConfigurationProps) => (
  <Table size="small">
    <TableBody>
      <TableRow>
        <TableCell>{i18n('common.name')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.name}
          </Typography>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell>{i18n('newE.target.kind')}</TableCell>
        <TableCell>
          <Typography variant="body2" color="textSecondary">
            {t.templateType}
          </Typography>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell>{i18n('newW.node.deadline')}</TableCell>
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
        {i18n('newE.steps.basic')}
      </Typography>
      <Table size="small">
        <TableBody>
          <TableRow>
            <TableCell>{i18n('common.name')}</TableCell>
            <TableCell>
              <Typography variant="body2" color="textSecondary">
                {t.name}
              </Typography>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <Typography variant="subtitle2" gutterBottom>
        {i18n('newW.node.container.title')}
      </Typography>
      <Table size="small">
        <TableRow>
          <TableCell>{i18n('common.name')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {container.name}
            </Typography>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>{i18n('newW.node.container.image')}</TableCell>
          <TableCell>
            <Typography variant="body2" color="textSecondary">
              {container.image}
            </Typography>
          </TableCell>
        </TableRow>
        {container.command && (
          <TableRow>
            <TableCell>{i18n('newW.node.container.command')}</TableCell>
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
        {i18n('newW.node.conditionalBranches.title')}
      </Typography>
      {t.conditionalBranches &&
        t.conditionalBranches.map((d: Branch, i: number) => (
          <Box key={i}>
            <Typography variant="subtitle2" gutterBottom>
              {i18n('newW.node.conditionalBranches.branch')} {i + 1}
            </Typography>
            <Table size="small">
              <TableRow>
                <TableCell>{i18n('newW.node.conditionalBranches.target')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {d.target}
                  </Typography>
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>{i18n('newW.node.conditionalBranches.expression')}</TableCell>
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
        return (
          <Box p={4.5}>
            <ObjectConfiguration config={t} inNode vertical />
          </Box>
        )
    }
  }

  return rendered()
}

export default NodeConfiguration
