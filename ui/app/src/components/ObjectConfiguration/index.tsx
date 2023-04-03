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
import { Grid, Table, TableBody, TableRow, Typography } from '@mui/material'
import { templateTypeToFieldName } from 'api/zz_generated.frontend.chaos-mesh'
import { TypesArchiveDetail, TypesExperimentDetail } from 'openapi/index.schemas'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import StatusLabel from 'components/StatusLabel'
import i18n from 'components/T'

import { format } from 'lib/luxon'

import { Experiment, Selector, TableCell } from './common'

type Config = TypesExperimentDetail | TypesArchiveDetail

interface ObjectConfigurationProps {
  config: Config
  inNode?: boolean
  inSchedule?: boolean
  inArchive?: boolean
  vertical?: boolean
}

const ObjectConfiguration: React.FC<ObjectConfigurationProps> = ({
  config,
  inNode,
  inSchedule,
  inArchive,
  vertical,
}) => {
  const { lang } = useStoreSelector((state) => state.settings)

  const spec: any = inNode ? config : config!.kube_object!.spec
  const experiment =
    inSchedule || inNode ? spec[templateTypeToFieldName(inSchedule ? spec.type : (config as any).templateType)] : spec

  return (
    <>
      {!inNode && (
        <Space direction="row" mb={3}>
          <Typography>{config.name}</Typography>

          {!inArchive && <StatusLabel status={(config as any).status} />}
        </Space>
      )}

      <Grid container spacing={vertical ? 3 : 0}>
        {!inNode && (
          <Grid item xs={vertical ? 12 : 3}>
            <Typography variant="subtitle2" gutterBottom>
              {i18n('newE.steps.basic')}
            </Typography>

            <Table size="small">
              <TableBody>
                <TableRow>
                  <TableCell>{i18n('k8s.namespace')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {config.namespace}
                    </Typography>
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>{i18n('common.uuid')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {config.uid}
                    </Typography>
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>{i18n('table.created')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {format(config.created_at!, lang)}
                    </Typography>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </Grid>
        )}

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {i18n('newE.steps.scope')}
          </Typography>

          {(inNode
            ? (config as any).templateType !== 'PhysicalMachineChaos'
            : inSchedule
            ? spec.type !== 'PhysicalMachineChaos'
            : (config.kind as any) !== 'PhysicalMachineChaos') && <Selector data={experiment.selector} />}

          {(inNode
            ? (config as any).templateType === 'PhysicalMachineChaos'
            : inSchedule
            ? spec.type === 'PhysicalMachineChaos'
            : (config.kind as any) === 'PhysicalMachineChaos') && (
            <Table>
              <TableRow>
                <TableCell>{i18n('physic.address')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {inNode
                      ? (config as any).physicalmachineChaos.address
                      : inSchedule
                      ? spec.physicalmachineChaos.address
                      : spec.address}
                  </Typography>
                </TableCell>
              </TableRow>
            </Table>
          )}
        </Grid>

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {i18n('experiments.single')}
          </Typography>

          <Experiment
            kind={inNode ? (config as any).templateType : inSchedule ? spec.type : config.kind}
            data={experiment}
          />
        </Grid>

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {i18n('newE.steps.run')}
          </Typography>

          <Table size="small">
            <TableBody>
              {!inSchedule && (
                <TableRow>
                  <TableCell>{i18n(inNode ? 'newW.node.deadline' : 'common.duration')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {inNode ? (config as any).deadline : spec.duration ? spec.duration : i18n('newE.run.continuous')}
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
              {inSchedule && (
                <>
                  <TableRow>
                    <TableCell>{i18n('schedules.single')}</TableCell>
                    <TableCell>
                      <Typography variant="body2" color="textSecondary">
                        {spec.schedule}
                      </Typography>
                    </TableCell>
                  </TableRow>
                  {spec.historyLimit && (
                    <TableRow>
                      <TableCell>{i18n('newS.basic.historyLimit')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.historyLimit}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                  {spec.concurrencyPolicy && (
                    <TableRow>
                      <TableCell>{i18n('newS.basic.concurrencyPolicy')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.concurrencyPolicy}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                  {spec.startingDeadlineSeconds && (
                    <TableRow>
                      <TableCell>{i18n('newS.basic.startingDeadlineSeconds')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.startingDeadlineSeconds}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </>
              )}
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </>
  )
}

export default ObjectConfiguration
